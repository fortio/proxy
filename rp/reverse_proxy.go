package rp

import (
	"crypto/tls"
	"flag"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
	"fortio.org/proxy/config"
	"golang.org/x/net/http2"
)

var (
	configs  = dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	h2Target = flag.Bool("h2", false, "Whether destinations support h2c prior knowledge")
)

// GetRoutes gets the current routes from the dynamic flag routes.json as object (deserialized).
func GetRoutes() []config.Route {
	routes := configs.Get().(*[]config.Route)
	return *routes
}

func setDestination(req *http.Request, url *url.URL) {
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
}

// Director is the object used by the ReverseProxy to pick the route/destination.
func Director(req *http.Request) {
	routes := GetRoutes()
	log.LogVf("Directing %+v", req)
	for _, route := range routes {
		log.LogVf("Evaluating req %q vs route %q and path %q vs prefix %q for dest %s",
			req.Host, route.Host, req.URL.Path, route.Prefix, route.Destination.URL.String())
		if route.MatchServerReq(req) {
			fhttp.LogRequest(req, route.Destination.Str)
			setDestination(req, &route.Destination.URL)
			return
		}
	}
}

// PrintRoutes prints the current value of the routes config (dflag).
func PrintRoutes() {
	if !log.Log(log.Info) {
		return
	}
	log.Printf("Initial Routes (routes.json dynamic flag):")
	for _, r := range GetRoutes() {
		log.Printf("host %q\t prefix %q\t -> %s", r.Host, r.Prefix, r.Destination.URL.String())
	}
}

// ReverseProxy returns a new reverse proxy which will route based on the config/dflags.
func ReverseProxy() *httputil.ReverseProxy {
	PrintRoutes()

	revp := httputil.ReverseProxy{Director: Director}

	// TODO: make h2c vs regular client more dynamic based on route config instead of all or nothing
	// (or maybe some day it will just ge the default behavior of the base http client)
	if *h2Target {
		revp.Transport = &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	}
	return &revp
}
