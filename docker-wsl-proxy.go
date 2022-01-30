package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/daemon/listeners"
	"github.com/docker/go-connections/sockets"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	// var wslDistro = flag.String("wsl-distro", "stevedore", "")
	var socketGroup = flag.String("g", "docker-users", "")

	const listenProto = "npipe"
	const listenAddr = "//./pipe/docker_engine_proxy"

	// TODO: we need to connect to docker running within WSL2.
	const connectProto = "npipe"
	const connectAddr = "//./pipe/docker_engine_linux"

	flag.Parse()

	proxyTarget := url.URL{
		Scheme: "http",
		Host:   "docker",
	}

	var err error

	proxy := httputil.NewSingleHostReverseProxy(&proxyTarget)
	proxy.Director = proxyDirector(proxy.Director)

	proxy.Transport, err = createTransport(connectProto, connectAddr)
	if err != nil {
		panic(err)
	}

	err = serve(listenProto, listenAddr, *socketGroup, proxy)
	if err != nil {
		panic(err)
	}
}

func createTransport(proto string, addr string) (http.RoundTripper, error) {
	transport := &http.Transport{}
	return transport, sockets.ConfigureTransport(transport, proto, addr)
}

func serve(proto string, addr string, socketGroup string, proxy *httputil.ReverseProxy) error {
	ls, err := listeners.Init(proto, addr, socketGroup, nil)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:    addr,
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
