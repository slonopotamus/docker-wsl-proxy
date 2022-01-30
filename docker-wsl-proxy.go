package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Microsoft/go-winio"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func createListener(network, addr, socketGroup string) (net.Listener, error) {
	switch network {
	case "npipe":
		// allow Administrators and SYSTEM, plus whatever additional users or groups were specified
		sddl := "D:P(A;;GA;;;BA)(A;;GA;;;SY)"
		if socketGroup != "" {
			for _, g := range strings.Split(socketGroup, ",") {
				sid, err := winio.LookupSidByName(g)
				if err != nil {
					return nil, err
				}
				sddl += fmt.Sprintf("(A;;GRGW;;;%s)", sid)
			}
		}

		c := winio.PipeConfig{
			SecurityDescriptor: sddl,
			MessageMode:        true,  // Use message mode so that CloseWrite() is supported
			InputBufferSize:    65536, // Use 64KB buffers to improve performance
			OutputBufferSize:   65536,
		}
		return winio.ListenPipe(addr, &c)

	default:
		return net.Listen(network, addr)
	}
}

func createTransport(network, addr string) *http.Transport {
	transport := http.Transport{}

	switch network {
	case "npipe":
		// No need for compression in local communications.
		transport.DisableCompression = true
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return winio.DialPipeContext(ctx, addr)
		}
	}

	return &transport
}

func main() {
	// var wslDistro = flag.String("wsl-distro", "stevedore", "")
	var socketGroup = flag.String("g", "docker-users", "")

	flag.Parse()

	proxyTarget := url.URL{
		Scheme: "http",
		Host:   "docker",
	}

	proxy := httputil.NewSingleHostReverseProxy(&proxyTarget)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		if strings.HasSuffix(req.RequestURI, "/containers/create") {
			// TODO: rewrite mount paths
			origDirector(req)
		} else {
			origDirector(req)
		}
	}

	// TODO: we need to connect to docker running within WSL2.
	proxy.Transport = createTransport("npipe", "//./pipe/docker_engine_linux")

	listener, err := createListener("npipe", "//./pipe/docker_engine_proxy", *socketGroup)
	if err != nil {
		panic(err)
	}

	httpServer := http.Server{
		Handler: proxy,
	}

	err = httpServer.Serve(listener)
	if err != nil {
		panic(err)
	}
}
