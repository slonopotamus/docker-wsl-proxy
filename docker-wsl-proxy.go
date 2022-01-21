package main

import (
	"context"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/server"
	"github.com/docker/docker/api/server/router"
	"github.com/docker/docker/api/server/router/checkpoint"
	containerrouter "github.com/docker/docker/api/server/router/container"
	distributionrouter "github.com/docker/docker/api/server/router/distribution"
	imagerouter "github.com/docker/docker/api/server/router/image"
	pluginrouter "github.com/docker/docker/api/server/router/plugin"
	"github.com/docker/docker/api/server/router/session"
	swarmrouter "github.com/docker/docker/api/server/router/swarm"
	"github.com/docker/docker/api/server/router/volume"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
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
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExecCreate(name string, config *types.ExecConfig) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExecInspect(id string) (*backend.ExecInspect, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExecResize(name string, height, width int) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExecStart(ctx context.Context, name string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ExecExists(name string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerArchivePath(name string, path string) (content io.ReadCloser, stat *types.ContainerPathStat, err error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerCopy(name string, res string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExport(name string, out io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerExtractToDir(name, path string, copyUIDGID, noOverwriteDirNonDir bool, content io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerStatPath(name string, path string) (stat *types.ContainerPathStat, err error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerCreate(config types.ContainerCreateConfig) (container.ContainerCreateCreatedBody, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerKill(name string, sig uint64) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerPause(name string) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerRename(oldName, newName string) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerResize(name string, height, width int) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerRestart(name string, seconds *int) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerRm(name string, config *types.ContainerRmConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerStart(name string, hostConfig *container.HostConfig, checkpoint string, checkpointDir string) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerStop(name string, seconds *int) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerUnpause(name string) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerUpdate(name string, hostConfig *container.HostConfig) (container.ContainerUpdateOKBody, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerWait(ctx context.Context, name string, condition containerpkg.WaitCondition) (<-chan containerpkg.StateStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerChanges(name string) ([]archive.Change, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerInspect(name string, size bool, version string) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerLogs(ctx context.Context, name string, config *types.ContainerLogsOptions) (msgs <-chan *backend.LogMessage, tty bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerStats(ctx context.Context, name string, config *backend.ContainerStatsConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerTop(name string, psArgs string) (*container.ContainerTopOKBody, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) Containers(config *types.ContainerListOptions) ([]*types.Container, error) {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainerAttach(name string, c *backend.ContainerAttachConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p ContainerProxy) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (*types.ContainersPruneReport, error) {
	//TODO implement me
	panic("implement me")
}

type ImageProxy struct {
	daemon ProxyDaemon
}

func (p ImageProxy) ImageDelete(imageRef string, force, prune bool) ([]types.ImageDeleteResponseItem, error) {
	return p.daemon.client.ImageRemove(p.daemon.ctx, imageRef, types.ImageRemoveOptions{Force: force, PruneChildren: prune})
}

func (p ImageProxy) ImageHistory(imageName string) ([]*image.HistoryResponseItem, error) {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) Images(imageFilters filters.Args, all bool, withExtraAttrs bool) ([]*types.ImageSummary, error) {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) LookupImage(name string) (*types.ImageInspect, error) {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) TagImage(imageName, repository, tag string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) ImagesPrune(ctx context.Context, pruneFilters filters.Args) (*types.ImagesPruneReport, error) {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) LoadImage(inTar io.ReadCloser, outStream io.Writer, quiet bool) error {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) ImportImage(src string, repository, platform string, tag string, msg string, inConfig io.ReadCloser, outStream io.Writer, changes []string) error {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) ExportImage(names []string, outStream io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) PullImage(ctx context.Context, image, tag string, platform *v1.Platform, metaHeaders map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) PushImage(ctx context.Context, image, tag string, metaHeaders map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (p ImageProxy) SearchRegistryForImages(ctx context.Context, filtersArgs string, term string, limit int, authConfig *types.AuthConfig, metaHeaders map[string][]string) (*registry.SearchResults, error) {
	//TODO implement me
	panic("implement me")
}

type VolumeProxy struct {
	daemon ProxyDaemon
}

func (p VolumeProxy) List(ctx context.Context, filter filters.Args) ([]*types.Volume, []string, error) {
	//TODO implement me
	panic("implement me")
}

func (p VolumeProxy) Get(ctx context.Context, name string, opts ...volumeopts.GetOption) (*types.Volume, error) {
	//TODO implement me
	panic("implement me")
}

func (p VolumeProxy) Create(ctx context.Context, name, driverName string, opts ...volumeopts.CreateOption) (*types.Volume, error) {
	//TODO implement me
	panic("implement me")
}

func (p VolumeProxy) Remove(ctx context.Context, name string, opts ...volumeopts.RemoveOption) error {
	//TODO implement me
	panic("implement me")
}

func (p VolumeProxy) Prune(ctx context.Context, pruneFilters filters.Args) (*types.VolumesPruneReport, error) {
	//TODO implement me
	panic("implement me")
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

func (p SwarmProxy) Update(u uint64, spec swarm.Spec, flags swarm.UpdateFlags) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetUnlockKey() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) UnlockSwarm(req swarm.UnlockRequest) error {
	return p.daemon.client.SwarmUnlock(p.daemon.ctx, req)
}

func (p SwarmProxy) GetServices(options types.ServiceListOptions) ([]swarm.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetService(idOrName string, insertDefaults bool) (swarm.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) CreateService(spec swarm.ServiceSpec, s string, b bool) (*types.ServiceCreateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) UpdateService(s string, u uint64, spec swarm.ServiceSpec, options types.ServiceUpdateOptions, b bool) (*types.ServiceUpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) RemoveService(s string) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) ServiceLogs(ctx context.Context, selector *backend.LogSelector, options *types.ContainerLogsOptions) (<-chan *backend.LogMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetNodes(options types.NodeListOptions) ([]swarm.Node, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetNode(s string) (swarm.Node, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) UpdateNode(s string, u uint64, spec swarm.NodeSpec) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) RemoveNode(s string, b bool) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetTasks(options types.TaskListOptions) ([]swarm.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetTask(s string) (swarm.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetSecrets(opts types.SecretListOptions) ([]swarm.Secret, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) CreateSecret(s swarm.SecretSpec) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) RemoveSecret(idOrName string) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetSecret(id string) (swarm.Secret, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) UpdateSecret(idOrName string, version uint64, spec swarm.SecretSpec) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetConfigs(opts types.ConfigListOptions) ([]swarm.Config, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) CreateConfig(s swarm.ConfigSpec) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) RemoveConfig(id string) error {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) GetConfig(id string) (swarm.Config, error) {
	//TODO implement me
	panic("implement me")
}

func (p SwarmProxy) UpdateConfig(idOrName string, version uint64, spec swarm.ConfigSpec) error {
	//TODO implement me
	panic("implement me")
}

type PluginProxy struct {
	daemon ProxyDaemon
}

func (p PluginProxy) Disable(name string, config *types.PluginDisableConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Enable(name string, config *types.PluginEnableConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) List(args filters.Args) ([]types.Plugin, error) {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Inspect(name string) (*types.Plugin, error) {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Remove(name string, config *types.PluginRmConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) Set(name string, args []string) error {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (p PluginProxy) CreateFromContext(ctx context.Context, tarCtx io.ReadCloser, options *types.PluginCreateOptions) error {
	//TODO implement me
	panic("implement me")
}

type DistributionProxy struct {
	daemon ProxyDaemon
}

func (p DistributionProxy) GetRepository(ctx context.Context, named reference.Named, config *types.AuthConfig) (distribution.Repository, bool, error) {
	//TODO implement me
	panic("implement me")
}

func main() {
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
		checkpoint.NewRouter(CheckpointProxy{daemon: daemon}, decoder),
		containerrouter.NewRouter(ContainerProxy{daemon: daemon}, decoder, false),
		imagerouter.NewRouter(ImageProxy{daemon: daemon}),
		//system.NewRouter(daemon, opts.cluster, opts.buildkit, opts.features),
		volume.NewRouter(VolumeProxy{daemon: daemon}),
		//build.NewRouter(opts.buildBackend, opts.daemon, opts.features),
		session.NewRouter(SessionProxy{daemon: daemon}),
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
