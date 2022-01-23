package main

import (
	"context"
	"flag"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/server"
	"github.com/docker/docker/api/server/router"
	checkpointrouter "github.com/docker/docker/api/server/router/checkpoint"
	containerrouter "github.com/docker/docker/api/server/router/container"
	distributionrouter "github.com/docker/docker/api/server/router/distribution"
	imagerouter "github.com/docker/docker/api/server/router/image"
	pluginrouter "github.com/docker/docker/api/server/router/plugin"
	sessionrouter "github.com/docker/docker/api/server/router/session"
	swarmrouter "github.com/docker/docker/api/server/router/swarm"
	volumerouter "github.com/docker/docker/api/server/router/volume"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	containerpkg "github.com/docker/docker/container"
	"github.com/docker/docker/daemon/listeners"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/docker/docker/plugin"
	"github.com/docker/docker/runconfig"
	volumeopts "github.com/docker/docker/volume/service/opts"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ProxyDaemon struct {
	ctx    context.Context
	client *client.Client
}

type CheckpointProxy struct {
	daemon ProxyDaemon
}

func (p CheckpointProxy) CheckpointCreate(container string, config types.CheckpointCreateOptions) error {
	return p.daemon.client.CheckpointCreate(p.daemon.ctx, container, config)
}

func (p CheckpointProxy) CheckpointDelete(container string, config types.CheckpointDeleteOptions) error {
	return p.daemon.client.CheckpointDelete(p.daemon.ctx, container, config)
}

func (p CheckpointProxy) CheckpointList(container string, config types.CheckpointListOptions) ([]types.Checkpoint, error) {
	return p.daemon.client.CheckpointList(p.daemon.ctx, container, config)
}

type ContainerProxy struct {
	daemon ProxyDaemon
}

func (p ContainerProxy) CreateImageFromContainer(name string, config *backend.CreateImageConfig) (imageID string, err error) {
	ref := [...]string{config.Repo, config.Tag}
	options := types.ContainerCommitOptions{
		Reference: strings.Join(ref[:], ":"),
		Comment:   config.Comment,
		Author:    config.Author,
		Changes:   config.Changes,
		Pause:     config.Pause,
		Config:    config.Config,
	}
	response, err := p.daemon.client.ContainerCommit(p.daemon.ctx, name, options)
	return response.ID, err
}

func (p ContainerProxy) ContainerExecCreate(name string, config *types.ExecConfig) (string, error) {
	response, err := p.daemon.client.ContainerExecCreate(p.daemon.ctx, name, *config)
	return response.ID, err
}

func (p ContainerProxy) ContainerExecInspect(id string) (*backend.ExecInspect, error) {
	inspect, err := p.daemon.client.ContainerExecInspect(p.daemon.ctx, id)
	return &backend.ExecInspect{
		ID:          inspect.ExecID,
		Running:     inspect.Running,
		ExitCode:    &inspect.ExitCode,
		ContainerID: inspect.ContainerID,
		Pid:         inspect.Pid,
	}, err
}

func (p ContainerProxy) ContainerExecResize(name string, height, width int) error {
	options := types.ResizeOptions{
		Height: uint(height),
		Width:  uint(width),
	}
	return p.daemon.client.ContainerExecResize(p.daemon.ctx, name, options)
}

func (p ContainerProxy) ContainerExecStart(ctx context.Context, name string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ExecExists(name string) (bool, error) {
	_, err := p.daemon.client.ContainerExecInspect(p.daemon.ctx, name)
	return true, err
}

func (p ContainerProxy) ContainerArchivePath(name string, path string) (content io.ReadCloser, stat *types.ContainerPathStat, err error) {
	content, pathStat, err := p.daemon.client.CopyFromContainer(p.daemon.ctx, name, path)
	return content, &pathStat, err
}

func (p ContainerProxy) ContainerCopy(name string, resource string) (io.ReadCloser, error) {
	data, _, err := p.daemon.client.CopyFromContainer(p.daemon.ctx, name, resource)
	return data, err
}

func (p ContainerProxy) ContainerExport(name string, out io.Writer) error {
	body, err := p.daemon.client.ContainerExport(p.daemon.ctx, name)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(out, body)
	return err
}

func (p ContainerProxy) ContainerExtractToDir(name, path string, copyUIDGID, noOverwriteDirNonDir bool, content io.Reader) error {
	options := types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: !noOverwriteDirNonDir,
		CopyUIDGID:                copyUIDGID,
	}
	return p.daemon.client.CopyToContainer(p.daemon.ctx, name, path, content, options)
}

func (p ContainerProxy) ContainerStatPath(name string, path string) (stat *types.ContainerPathStat, err error) {
	statPath, err := p.daemon.client.ContainerStatPath(p.daemon.ctx, name, path)
	return &statPath, err
}

func (p ContainerProxy) ContainerCreate(config types.ContainerCreateConfig) (container.ContainerCreateCreatedBody, error) {
	return p.daemon.client.ContainerCreate(p.daemon.ctx, config.Config, config.HostConfig, config.NetworkingConfig, config.Platform, config.Name)
}

func (p ContainerProxy) ContainerKill(name string, sig uint64) error {
	return p.daemon.client.ContainerKill(p.daemon.ctx, name, strconv.FormatUint(sig, 10))
}

func (p ContainerProxy) ContainerPause(name string) error {
	return p.daemon.client.ContainerPause(p.daemon.ctx, name)
}

func (p ContainerProxy) ContainerRename(oldName, newName string) error {
	return p.daemon.client.ContainerRename(p.daemon.ctx, oldName, newName)
}

func (p ContainerProxy) ContainerResize(name string, height, width int) error {
	options := types.ResizeOptions{
		Height: uint(height),
		Width:  uint(width),
	}
	return p.daemon.client.ContainerResize(p.daemon.ctx, name, options)
}

func (p ContainerProxy) ContainerRestart(name string, seconds *int) error {
	timeout := time.Duration(*seconds) * time.Second
	return p.daemon.client.ContainerRestart(p.daemon.ctx, name, &timeout)
}

func (p ContainerProxy) ContainerRm(name string, config *types.ContainerRmConfig) error {
	options := types.ContainerRemoveOptions{
		RemoveVolumes: config.RemoveVolume,
		RemoveLinks:   config.RemoveLink,
		Force:         config.ForceRemove,
	}
	return p.daemon.client.ContainerRemove(p.daemon.ctx, name, options)
}

func (p ContainerProxy) ContainerStart(name string, _ *container.HostConfig, checkpoint string, checkpointDir string) error {
	options := types.ContainerStartOptions{
		CheckpointID:  checkpoint,
		CheckpointDir: checkpointDir,
	}
	return p.daemon.client.ContainerStart(p.daemon.ctx, name, options)
}

func (p ContainerProxy) ContainerStop(name string, seconds *int) error {
	timeout := time.Duration(*seconds) * time.Second
	return p.daemon.client.ContainerStop(p.daemon.ctx, name, &timeout)
}

func (p ContainerProxy) ContainerUnpause(name string) error {
	return p.daemon.client.ContainerUnpause(p.daemon.ctx, name)
}

func (p ContainerProxy) ContainerUpdate(name string, hostConfig *container.HostConfig) (container.ContainerUpdateOKBody, error) {
	config := container.UpdateConfig{
		Resources:     hostConfig.Resources,
		RestartPolicy: hostConfig.RestartPolicy,
	}
	return p.daemon.client.ContainerUpdate(p.daemon.ctx, name, config)
}

func (p ContainerProxy) ContainerWait(ctx context.Context, name string, condition containerpkg.WaitCondition) (<-chan containerpkg.StateStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerChanges(name string) ([]archive.Change, error) {
	diff, err := p.daemon.client.ContainerDiff(p.daemon.ctx, name)
	result := make([]archive.Change, len(diff))
	for index, item := range diff {
		result[index] = archive.Change{
			Path: item.Path,
			Kind: archive.ChangeType(item.Kind),
		}
	}

	return result, err
}

func (p ContainerProxy) ContainerInspect(name string, _ bool, _ string) (interface{}, error) {
	return p.daemon.client.ContainerInspect(p.daemon.ctx, name)
}

func (p ContainerProxy) ContainerLogs(ctx context.Context, name string, config *types.ContainerLogsOptions) (msgs <-chan *backend.LogMessage, tty bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerStats(ctx context.Context, name string, config *backend.ContainerStatsConfig) error {
	var stats types.ContainerStats
	var err error

	if config.OneShot {
		stats, err = p.daemon.client.ContainerStatsOneShot(ctx, name)
	} else {
		stats, err = p.daemon.client.ContainerStats(ctx, name, config.Stream)
	}
	if err != nil {
		return err
	}
	defer stats.Body.Close()
	_, err = io.Copy(config.OutStream, stats.Body)
	return err
}

func (p ContainerProxy) ContainerTop(name string, psArgs string) (*container.ContainerTopOKBody, error) {
	arguments := [...]string{psArgs}
	body, err := p.daemon.client.ContainerTop(p.daemon.ctx, name, arguments[:])
	return &body, err
}

func (p ContainerProxy) Containers(config *types.ContainerListOptions) ([]*types.Container, error) {
	containers, err := p.daemon.client.ContainerList(p.daemon.ctx, *config)

	result := make([]*types.Container, len(containers))
	for index, c := range containers {
		result[index] = &c
	}

	return result, err
}

func (p ContainerProxy) ContainerAttach(name string, config *backend.ContainerAttachConfig) error {
	options := types.ContainerAttachOptions{
		Stream:     config.Stream,
		Stdin:      config.UseStdin,
		Stdout:     config.UseStdout,
		Stderr:     config.UseStderr,
		DetachKeys: config.DetachKeys,
		Logs:       config.Logs,
	}
	response, err := p.daemon.client.ContainerAttach(p.daemon.ctx, name, options)
	response.Close()
	return err
}

func (p ContainerProxy) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (*types.ContainersPruneReport, error) {
	report, err := p.daemon.client.ContainersPrune(ctx, pruneFilters)
	return &report, err
}

type ImageProxy struct {
	daemon ProxyDaemon
}

func (p ImageProxy) ImageDelete(imageRef string, force, prune bool) ([]types.ImageDeleteResponseItem, error) {
	return p.daemon.client.ImageRemove(p.daemon.ctx, imageRef, types.ImageRemoveOptions{Force: force, PruneChildren: prune})
}

func (p ImageProxy) ImageHistory(imageName string) ([]*image.HistoryResponseItem, error) {
	history, err := p.daemon.client.ImageHistory(p.daemon.ctx, imageName)

	result := make([]*image.HistoryResponseItem, len(history))
	for index, item := range history {
		result[index] = &item
	}

	return result, err
}

func (p ImageProxy) Images(imageFilters filters.Args, all bool, _ bool) ([]*types.ImageSummary, error) {
	options := types.ImageListOptions{
		All:     all,
		Filters: imageFilters,
	}

	summaries, err := p.daemon.client.ImageList(p.daemon.ctx, options)

	result := make([]*types.ImageSummary, len(summaries))
	for index, summary := range summaries {
		result[index] = &summary
	}

	return result, err
}

func (p ImageProxy) LookupImage(name string) (*types.ImageInspect, error) {
	inspect, _, err := p.daemon.client.ImageInspectWithRaw(p.daemon.ctx, name)
	return &inspect, err
}

func (p ImageProxy) TagImage(imageName, repository, tag string) (string, error) {
	ref := [...]string{repository, tag}
	refstr := strings.Join(ref[:], ":")
	err := p.daemon.client.ImageTag(p.daemon.ctx, imageName, refstr)
	return refstr, err
}

func (p ImageProxy) ImagesPrune(ctx context.Context, pruneFilters filters.Args) (*types.ImagesPruneReport, error) {
	report, err := p.daemon.client.ImagesPrune(ctx, pruneFilters)
	return &report, err
}

func (p ImageProxy) LoadImage(inTar io.ReadCloser, outStream io.Writer, quiet bool) error {
	response, err := p.daemon.client.ImageLoad(p.daemon.ctx, inTar, quiet)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, err = io.Copy(outStream, response.Body)
	return err
}

func (p ImageProxy) ImportImage(src string, repository, platform string, tag string, message string, inConfig io.ReadCloser, outStream io.Writer, changes []string) error {
	ref := [...]string{repository, tag}

	importSource := types.ImageImportSource{
		Source:     inConfig,
		SourceName: src,
	}

	options := types.ImageImportOptions{
		Tag:      tag,
		Message:  message,
		Changes:  changes,
		Platform: platform,
	}

	response, err := p.daemon.client.ImageImport(p.daemon.ctx, importSource, strings.Join(ref[:], ":"), options)
	if err != nil {
		return nil
	}
	defer response.Close()
	_, err = io.Copy(outStream, response)
	return err
}

func (p ImageProxy) ExportImage(names []string, outStream io.Writer) error {
	body, err := p.daemon.client.ImageSave(p.daemon.ctx, names)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(outStream, body)
	return err
}

func (p ImageProxy) PullImage(ctx context.Context, image, tag string, platform *v1.Platform, _ map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error {
	ref := [...]string{image, tag}

	platformStr := ""
	if platform != nil {
		platformStr = platforms.Format(*platform)
	}

	options := types.ImagePullOptions{
		// TODO implement me
		RegistryAuth:  "",
		PrivilegeFunc: nil,
		Platform:      platformStr,
	}

	body, err := p.daemon.client.ImagePull(ctx, strings.Join(ref[:], ":"), options)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(outStream, body)
	return err
}

func (p ImageProxy) PushImage(ctx context.Context, image, tag string, _ map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error {
	ref := [...]string{image, tag}

	options := types.ImagePushOptions{
		// TODO implement me
		RegistryAuth:  "",
		PrivilegeFunc: nil,
	}

	body, err := p.daemon.client.ImagePush(ctx, strings.Join(ref[:], ":"), options)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(outStream, body)
	return err
}

func (p ImageProxy) SearchRegistryForImages(ctx context.Context, filtersArgs string, term string, limit int, authConfig *types.AuthConfig, _ map[string][]string) (*registry.SearchResults, error) {
	f, err := filters.FromJSON(filtersArgs)
	if err != nil {
		return nil, err
	}

	options := types.ImageSearchOptions{
		// TODO implement me
		RegistryAuth:  "",
		PrivilegeFunc: nil,
		Filters:       f,
		Limit:         limit,
	}

	results, err := p.daemon.client.ImageSearch(ctx, term, options)
	return &registry.SearchResults{
		Query:      "",
		NumResults: len(results),
		Results:    results,
	}, err
}

type VolumeProxy struct {
	daemon ProxyDaemon
}

func (p VolumeProxy) List(ctx context.Context, filter filters.Args) ([]*types.Volume, []string, error) {
	volumeList, err := p.daemon.client.VolumeList(ctx, filter)
	return volumeList.Volumes, volumeList.Warnings, err
}

func (p VolumeProxy) Get(ctx context.Context, name string, _ ...volumeopts.GetOption) (*types.Volume, error) {
	vol, err := p.daemon.client.VolumeInspect(ctx, name)
	return &vol, err
}

func (p VolumeProxy) Create(ctx context.Context, name, driverName string, opts ...volumeopts.CreateOption) (*types.Volume, error) {
	config := volumeopts.CreateConfig{}

	for _, opt := range opts {
		opt(&config)
	}

	body := volume.VolumeCreateBody{
		Driver:     driverName,
		DriverOpts: config.Options,
		Labels:     config.Labels,
		Name:       name,
	}

	response, err := p.daemon.client.VolumeCreate(ctx, body)
	return &response, err
}

func (p VolumeProxy) Remove(ctx context.Context, name string, opts ...volumeopts.RemoveOption) error {
	var config volumeopts.RemoveConfig

	for _, opt := range opts {
		opt(&config)
	}

	return p.daemon.client.VolumeRemove(ctx, name, config.PurgeOnError)
}

func (p VolumeProxy) Prune(ctx context.Context, pruneFilters filters.Args) (*types.VolumesPruneReport, error) {
	report, err := p.daemon.client.VolumesPrune(ctx, pruneFilters)
	return &report, err
}

type SessionProxy struct {
	daemon ProxyDaemon
}

func (p SessionProxy) HandleHTTPRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

type SwarmProxy struct {
	daemon ProxyDaemon
}

func (p SwarmProxy) Init(req swarm.InitRequest) (string, error) {
	return p.daemon.client.SwarmInit(p.daemon.ctx, req)
}

func (p SwarmProxy) Join(req swarm.JoinRequest) error {
	return p.daemon.client.SwarmJoin(p.daemon.ctx, req)
}

func (p SwarmProxy) Leave(force bool) error {
	return p.daemon.client.SwarmLeave(p.daemon.ctx, force)
}

func (p SwarmProxy) Inspect() (swarm.Swarm, error) {
	return p.daemon.client.SwarmInspect(p.daemon.ctx)
}

func (p SwarmProxy) Update(version uint64, spec swarm.Spec, flags swarm.UpdateFlags) error {
	return p.daemon.client.SwarmUpdate(p.daemon.ctx, swarm.Version{Index: version}, spec, flags)
}

func (p SwarmProxy) GetUnlockKey() (string, error) {
	response, err := p.daemon.client.SwarmGetUnlockKey(p.daemon.ctx)
	return response.UnlockKey, err
}

func (p SwarmProxy) UnlockSwarm(req swarm.UnlockRequest) error {
	return p.daemon.client.SwarmUnlock(p.daemon.ctx, req)
}

func (p SwarmProxy) GetServices(options types.ServiceListOptions) ([]swarm.Service, error) {
	return p.daemon.client.ServiceList(p.daemon.ctx, options)
}

func (p SwarmProxy) GetService(id string, insertDefaults bool) (swarm.Service, error) {
	options := types.ServiceInspectOptions{InsertDefaults: insertDefaults}
	service, _, err := p.daemon.client.ServiceInspectWithRaw(p.daemon.ctx, id, options)
	return service, err
}

func (p SwarmProxy) CreateService(spec swarm.ServiceSpec, encodedRegistryAuth string, queryRegistry bool) (*types.ServiceCreateResponse, error) {
	options := types.ServiceCreateOptions{
		EncodedRegistryAuth: encodedRegistryAuth,
		QueryRegistry:       queryRegistry,
	}
	response, err := p.daemon.client.ServiceCreate(p.daemon.ctx, spec, options)
	return &response, err
}

func (p SwarmProxy) UpdateService(serviceID string, version uint64, spec swarm.ServiceSpec, options types.ServiceUpdateOptions, _ bool) (*types.ServiceUpdateResponse, error) {
	response, err := p.daemon.client.ServiceUpdate(p.daemon.ctx, serviceID, swarm.Version{Index: version}, spec, options)
	return &response, err
}

func (p SwarmProxy) RemoveService(s string) error {
	return p.daemon.client.ServiceRemove(p.daemon.ctx, s)
}

func (p SwarmProxy) ServiceLogs(ctx context.Context, selector *backend.LogSelector, options *types.ContainerLogsOptions) (<-chan *backend.LogMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetNodes(options types.NodeListOptions) ([]swarm.Node, error) {
	return p.daemon.client.NodeList(p.daemon.ctx, options)
}

func (p SwarmProxy) GetNode(nodeID string) (swarm.Node, error) {
	node, _, err := p.daemon.client.NodeInspectWithRaw(p.daemon.ctx, nodeID)
	return node, err
}

func (p SwarmProxy) UpdateNode(nodeID string, version uint64, spec swarm.NodeSpec) error {
	return p.daemon.client.NodeUpdate(p.daemon.ctx, nodeID, swarm.Version{Index: version}, spec)
}

func (p SwarmProxy) RemoveNode(nodeID string, force bool) error {
	options := types.NodeRemoveOptions{
		Force: force,
	}
	return p.daemon.client.NodeRemove(p.daemon.ctx, nodeID, options)
}

func (p SwarmProxy) GetTasks(options types.TaskListOptions) ([]swarm.Task, error) {
	return p.daemon.client.TaskList(p.daemon.ctx, options)
}

func (p SwarmProxy) GetTask(taskID string) (swarm.Task, error) {
	task, _, err := p.daemon.client.TaskInspectWithRaw(p.daemon.ctx, taskID)
	return task, err
}

func (p SwarmProxy) GetSecrets(opts types.SecretListOptions) ([]swarm.Secret, error) {
	return p.daemon.client.SecretList(p.daemon.ctx, opts)
}

func (p SwarmProxy) CreateSecret(secret swarm.SecretSpec) (string, error) {
	response, err := p.daemon.client.SecretCreate(p.daemon.ctx, secret)
	return response.ID, err
}

func (p SwarmProxy) RemoveSecret(id string) error {
	return p.daemon.client.SecretRemove(p.daemon.ctx, id)
}

func (p SwarmProxy) GetSecret(id string) (swarm.Secret, error) {
	secret, _, err := p.daemon.client.SecretInspectWithRaw(p.daemon.ctx, id)
	return secret, err
}

func (p SwarmProxy) UpdateSecret(id string, version uint64, spec swarm.SecretSpec) error {
	return p.daemon.client.SecretUpdate(p.daemon.ctx, id, swarm.Version{Index: version}, spec)
}

func (p SwarmProxy) GetConfigs(options types.ConfigListOptions) ([]swarm.Config, error) {
	return p.daemon.client.ConfigList(p.daemon.ctx, options)
}

func (p SwarmProxy) CreateConfig(spec swarm.ConfigSpec) (string, error) {
	response, err := p.daemon.client.ConfigCreate(p.daemon.ctx, spec)
	return response.ID, err
}

func (p SwarmProxy) RemoveConfig(id string) error {
	return p.daemon.client.ConfigRemove(p.daemon.ctx, id)
}

func (p SwarmProxy) GetConfig(id string) (swarm.Config, error) {
	config, _, err := p.daemon.client.ConfigInspectWithRaw(p.daemon.ctx, id)
	return config, err
}

func (p SwarmProxy) UpdateConfig(id string, version uint64, spec swarm.ConfigSpec) error {
	return p.daemon.client.ConfigUpdate(p.daemon.ctx, id, swarm.Version{Index: version}, spec)
}

type PluginProxy struct {
	daemon ProxyDaemon
}

func (p PluginProxy) Disable(name string, config *types.PluginDisableConfig) error {
	return p.daemon.client.PluginDisable(p.daemon.ctx, name, types.PluginDisableOptions{Force: config.ForceDisable})
}

func (p PluginProxy) Enable(name string, config *types.PluginEnableConfig) error {
	return p.daemon.client.PluginEnable(p.daemon.ctx, name, types.PluginEnableOptions{Timeout: config.Timeout})
}

func (p PluginProxy) List(args filters.Args) ([]types.Plugin, error) {
	list, err := p.daemon.client.PluginList(p.daemon.ctx, args)
	result := make([]types.Plugin, len(list))
	for index, item := range list {
		result[index] = *item
	}
	return result, err
}

func (p PluginProxy) Inspect(name string) (*types.Plugin, error) {
	response, _, err := p.daemon.client.PluginInspectWithRaw(p.daemon.ctx, name)
	return response, err
}

func (p PluginProxy) Remove(name string, config *types.PluginRmConfig) error {
	return p.daemon.client.PluginRemove(p.daemon.ctx, name, types.PluginRemoveOptions{Force: config.ForceRemove})
}

func (p PluginProxy) Set(name string, args []string) error {
	return p.daemon.client.PluginSet(p.daemon.ctx, name, args)
}

func (p PluginProxy) Privileges(ctx context.Context, ref reference.Named, metaHeaders http.Header, authConfig *types.AuthConfig) (types.PluginPrivileges, error) {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Pull(ctx context.Context, ref reference.Named, name string, metaHeaders http.Header, authConfig *types.AuthConfig, privileges types.PluginPrivileges, outStream io.Writer, opts ...plugin.CreateOpt) error {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Push(ctx context.Context, name string, metaHeaders http.Header, authConfig *types.AuthConfig, outStream io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Upgrade(ctx context.Context, ref reference.Named, name string, metaHeaders http.Header, authConfig *types.AuthConfig, privileges types.PluginPrivileges, outStream io.Writer) error {
	options := types.PluginInstallOptions{
		// TODO implement me
	}

	body, err := p.daemon.client.PluginUpgrade(ctx, name, options)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(outStream, body)
	return err
}

func (p PluginProxy) CreateFromContext(ctx context.Context, tarCtx io.ReadCloser, options *types.PluginCreateOptions) error {
	return p.daemon.client.PluginCreate(ctx, tarCtx, *options)
}

type DistributionProxy struct {
	daemon ProxyDaemon
}

func (p DistributionProxy) GetRepository(ctx context.Context, named reference.Named, config *types.AuthConfig) (distribution.Repository, bool, error) {
	//TODO implement me
	panic("implement me")
}

var wslDistro = flag.String("wsl-distro", "stevedore", "")

func main() {
	flag.Parse()

	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	daemon := ProxyDaemon{client: c, ctx: context.Background()}
	decoder := runconfig.ContainerDecoder{
		GetSysInfo: func() *sysinfo.SysInfo {
			return sysinfo.New(true)
		},
	}

	serverConfig := server.Config{}
	srv := server.New(&serverConfig)

	routers := []router.Router{
		// we need to add the checkpoint router before the container router or the DELETE gets masked
		checkpointrouter.NewRouter(CheckpointProxy{daemon: daemon}, decoder),
		containerrouter.NewRouter(ContainerProxy{daemon: daemon}, decoder, false),
		imagerouter.NewRouter(ImageProxy{daemon: daemon}),
		//system.NewRouter(daemon, opts.cluster, opts.buildkit, opts.features),
		volumerouter.NewRouter(VolumeProxy{daemon: daemon}),
		//build.NewRouter(opts.buildBackend, opts.daemon, opts.features),
		sessionrouter.NewRouter(SessionProxy{daemon: daemon}),
		swarmrouter.NewRouter(SwarmProxy{daemon: daemon}),
		pluginrouter.NewRouter(PluginProxy{daemon: daemon}),
		distributionrouter.NewRouter(DistributionProxy{daemon: daemon}),
	}

	srv.InitRouter(routers...)

	//ls, err := listeners.Init("npipe", opts.DefaultNamedPipe, serverConfig.SocketGroup, serverConfig.TLSConfig)
	ls, err := listeners.Init("tcp", "127.0.0.1:12345", serverConfig.SocketGroup, serverConfig.TLSConfig)
	if err != nil {
		panic(err)
	}

	srv.Accept("0.0.0.0", ls...)

	serveAPIWait := make(chan error)
	go srv.Wait(serveAPIWait)
	err = <-serveAPIWait
	if err != nil {
		panic(err)
	}
}
