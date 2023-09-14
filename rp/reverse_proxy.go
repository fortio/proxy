package rp

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sort"
	"time"

	"fortio.org/dflag"
	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/version"
	"fortio.org/log"
	"fortio.org/proxy/config"
	"golang.org/x/net/http2"
)

var (
	configs   = dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	h2Target  = flag.Bool("h2", false, "Whether destinations support h2c prior knowledge")
	HostID    = dflag.DynString(flag.CommandLine, "hostid", "", "host id to show in debug-host output")
	startTime = time.Now()
	// optional fortio debug virtual host.
	DebugHost = dflag.DynString(flag.CommandLine, "debug-host", "",
		"`hostname` to serve echo debug info on if non-empty (ex: debug.fortio.org)")
)

// GetRoutes gets the current routes from the dynamic flag routes.json as object (deserialized).
func GetRoutes() []config.Route {
	routes := configs.Get().(*[]config.Route)
	return *routes
}

const noRouteMarker = "no-route"

// Rewrite is how incoming request are processed for the ReverseProxy
// to pick the route/destination.
func Rewrite(pr *httputil.ProxyRequest) {
	routes := GetRoutes()
	log.LogVf("RP rewrite %+v", pr)
	req := pr.In
	for _, route := range routes {
		log.LogVf("Evaluating req %q vs route %q and path %q vs prefix %q for dest %s",
			req.Host, route.Host, req.URL.Path, route.Prefix, route.Destination.URL.String())
		if route.MatchServerReq(req) {
			pr.SetXForwarded()
			//nolint:gosec // we return after this so there is only 1 pointer to the URL
			pr.SetURL(&route.Destination.URL)
			log.LogRequest(req, route.Destination.Str)
			return
		}
	}
	// No route matched, log and return 404.
	log.Errf("No route matched for %q %q", req.Host, req.URL.Path)
	pr.Out.URL.Scheme = noRouteMarker
}

// ErrorHandler is the error handler for the ReverseProxy. We use
// a Scheme marker to know that the error is just there was no route
// and treat that as 404 and everything else remains a 502.
func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if r.URL.Scheme == noRouteMarker {
		http.Error(w, "No route matched", http.StatusNotFound)
	} else {
		log.Errf("Proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
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

	revp := httputil.ReverseProxy{
		Rewrite:      Rewrite,
		ErrorHandler: ErrorHandler,
	}

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

// GzipDebugHandler is a handler wrapping SafeDebugHandler with optional gzip compression.
var GzipDebugHandler = fhttp.Gzip(http.HandlerFunc(SafeDebugHandler))

// DebugHandler is similar to Fortio's DebugHandler,
// it returns debug/useful info to http client.
// but doesn't have some of the extra sensitive info like env dump
// and host name or echo delay or header setting options.
func SafeDebugHandler(w http.ResponseWriter, r *http.Request) {
	log.LogRequest(r, "Debug")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	var buf bytes.Buffer
	buf.WriteString("Φορτίο version ")
	buf.WriteString(version.Long())
	buf.WriteString("\nDebug server")
	id := HostID.Get()
	if id != "" {
		buf.WriteString(" on ")
		buf.WriteString(id)
	}
	buf.WriteString(" up for ")
	buf.WriteString(fmt.Sprint(fhttp.RoundDuration(time.Since(startTime))))
	buf.WriteString("\nRequest from ")
	buf.WriteString(r.RemoteAddr)
	buf.WriteString(log.TLSInfo(r))
	buf.WriteString("\n\n")
	buf.WriteString(r.Method)
	buf.WriteByte(' ')
	buf.WriteString(r.URL.String())
	buf.WriteByte(' ')
	buf.WriteString(r.Proto)
	buf.WriteString("\n\nheaders:\n\n")
	// Host is removed from headers map and put here (!)
	buf.WriteString("Host: ")
	buf.WriteString(r.Host)

	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		buf.WriteByte('\n')
		buf.WriteString(name)
		buf.WriteString(": ")
		first := true
		headers := r.Header[name]
		for _, h := range headers {
			if !first {
				buf.WriteByte(',')
			}
			buf.WriteString(h)
			first = false
		}
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errf("Error reading %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteString("\n\nbody:\n\n")
	buf.WriteString(fhttp.DebugSummary(data, 512))
	buf.WriteByte('\n')
	if _, err = w.Write(buf.Bytes()); err != nil {
		log.Errf("Error writing response %v to %v", err, r.RemoteAddr)
	}
}

func DebugOnHostHandler(normalHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		debugHost := DebugHost.Get()
		if debugHost != "" && r.Host == debugHost && r.URL.Path != "/favicon.ico" {
			GzipDebugHandler.ServeHTTP(w, r)
		} else {
			normalHandler(w, r)
		}
	}
}
