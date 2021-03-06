package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/daemon/listeners"
	"github.com/docker/go-connections/sockets"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
	_ "unsafe"
)

type CmdConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (c CmdConn) Read(b []byte) (n int, err error) {
	return c.stdout.Read(b)
}

func (c CmdConn) Write(b []byte) (n int, err error) {
	return c.stdin.Write(b)
}

func (c CmdConn) Close() error {
	err := c.cmd.Process.Kill()
	_ = c.cmd.Wait()
	return err
}

func (c CmdConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (c CmdConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (c CmdConn) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (c CmdConn) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (c CmdConn) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func configureTransport(transport *http.Transport, connectString string) error {
	connectURL, err := url.Parse(connectString)
	if err != nil {
		return err
	}

	if connectURL.Scheme == "wsl" {
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			cmd := exec.Command("wsl.exe", "-d", connectURL.Host, "socat", fmt.Sprintf("UNIX-CONNECT:%s", connectURL.Path), "-")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return nil, err
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return nil, err
			}
			err = cmd.Start()
			if err != nil {
				return nil, err
			}
			return CmdConn{cmd, stdin, stdout}, nil
		}
		return nil
	}

	connectURL, err = client.ParseHostURL(connectString)
	if err != nil {
		return err
	}

	return sockets.ConfigureTransport(transport, connectURL.Scheme, connectURL.Host)
}

func main() {
	connectString := flag.String(
		"c", "wsl://stevedore/var/run/docker.sock",
		"address of docker daemon",
	)
	listenString := flag.String(
		"l", "npipe:////./pipe/docker_engine",
		"address where docker-wsl-proxy will listen to incoming connections",
	)
	socketGroup := flag.String(
		"g", "docker-users",
		"Windows group that will have access to docker-wsl-proxy",
	)

	flag.Parse()

	listenURL, err := client.ParseHostURL(*listenString)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "docker",
	})
	proxy.Director = proxyDirector(proxy.Director)
	proxy.ModifyResponse = proxyModifyResponse

	proxy.Transport, err = createTransport(*connectString)
	if err != nil {
		panic(err)
	}

	err = serve(listenURL, *socketGroup, proxy)
	if err != nil {
		panic(err)
	}
}

func createTransport(connectString string) (http.RoundTripper, error) {
	transport := &http.Transport{}
	return transport, configureTransport(transport, connectString)
}

func serve(listenURL *url.URL, socketGroup string, proxy *httputil.ReverseProxy) error {
	ls, err := listeners.Init(listenURL.Scheme, listenURL.Host, socketGroup, nil)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:    listenURL.Host,
		Handler: proxy,
	}

	return server.Serve(ls[0])
}

type configWrapper struct {
	*container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
	Platform         *specs.Platform
}

func proxyDirector(origDirector func(*http.Request)) func(req *http.Request) {
	return func(req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/containers/create") {
			err := handleContainerCreate(req)
			if err != nil {
				panic(err)
			}
			origDirector(req)
		} else {
			origDirector(req)
		}
	}
}

func handleContainerCreate(req *http.Request) error {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var config configWrapper
	if err = json.Unmarshal(buf, &config); err != nil {
		return err
	}

	for index, bind := range config.HostConfig.Binds {
		config.HostConfig.Binds[index] = rewriteBindToWSL(bind)
	}

	for _, m := range config.HostConfig.Mounts {
		if m.Type == mount.TypeBind {
			m.Source = rewriteBindToWSL(m.Source)
		}
	}

	buf, err = json.Marshal(config)
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(buf))
	req.ContentLength = int64(len(buf))

	return nil
}

var containerInspectRegexp = regexp.MustCompile(`/.*/containers/.*/json`)

func proxyModifyResponse(response *http.Response) (err error) {
	if containerInspectRegexp.MatchString(response.Request.URL.Path) && response.StatusCode == http.StatusOK {
		err = handleContainerInspect(response)
	}

	return err
}

func handleContainerInspect(response *http.Response) error {
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var containerJSON types.ContainerJSON
	if err = json.Unmarshal(buf, &containerJSON); err != nil {
		return err
	}

	for index, bind := range containerJSON.HostConfig.Binds {
		containerJSON.HostConfig.Binds[index] = rewriteBindToWindows(bind)
	}

	for _, m := range containerJSON.HostConfig.Mounts {
		if m.Type == mount.TypeBind {
			m.Source = rewriteBindSourceToWindows(m.Source)
		}
	}

	for _, m := range containerJSON.Mounts {
		if m.Type == mount.TypeBind {
			m.Source = rewriteBindSourceToWindows(m.Source)
		}
	}

	buf, err = json.Marshal(containerJSON)
	if err != nil {
		return err
	}

	response.Body = io.NopCloser(bytes.NewReader(buf))
	response.ContentLength = int64(len(buf))
	return err
}

func rewriteBindToWSL(s string) string {
	s = strings.ReplaceAll(s, `\`, "/")
	s = path.Clean(s)
	parts := strings.Split(s, ":")
	if strings.HasPrefix(parts[0], "/host_mnt/") {
		parts[0] = "/mnt/" + parts[0][10:]
		s = strings.Join(parts, ":")
	} else if len(parts[0]) == 1 && parts[0] != "/" {
		s = "/mnt/" + strings.ToLower(parts[0])
		s += strings.Join(parts[1:], ":")
	} else if len(parts[0]) > 2 && parts[0][0] == '/' && parts[0][2] == '/' {
		parts[0] = "/mnt/" + strings.ToLower(parts[0][1:2]) + parts[0][2:]
		s = strings.Join(parts, ":")
	}
	return s
}

func rewriteBindToWindows(bind string) string {
	parts := strings.Split(bind, ":")
	parts[0] = rewriteBindSourceToWindows(parts[0])
	return strings.Join(parts, ":")
}

func rewriteBindSourceToWindows(source string) string {
	source = path.Clean(source)

	if strings.HasPrefix(source, "/") {
		p := strings.Split(source[1:], "/")
		if len(p) > 1 && p[0] == "mnt" {
			p = p[1:]
			p[0] = strings.ToUpper(p[0]) + ":"
			source = strings.Join(p, `\`)
		}
	}

	return source
}
