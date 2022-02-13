package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/daemon/listeners"
	"github.com/docker/go-connections/sockets"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
	"time"
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
		if strings.HasSuffix(req.RequestURI, "/containers/create") {
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
	buf := make([]byte, req.ContentLength)
	_, err := io.ReadFull(req.Body, buf)
	if err != nil {
		return err
	}

	config := configWrapper{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return err
	}

	// TODO: rewrite config.HostConfig.Binds

	buf, err = json.Marshal(config)
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(buf))
	req.ContentLength = int64(len(buf))

	return nil
}
