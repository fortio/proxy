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
	"net/url"
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

func setDestination(req *http.Request, url *url.URL) {
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	// Horrible hack to workaround golang behavior with User-Agent: addition
	// same "fix" as https://github.com/golang/go/commit/6a6c1d9841a1957a2fd292df776ea920ae38ea00
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
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

// GzipDebugHandler is a handler wrapping SafeDebugHandler with optional gzip compression.
var GzipDebugHandler = fhttp.Gzip(http.HandlerFunc(SafeDebugHandler))

// DebugHandler is similar to Fortio's DebugHandler,
// it returns debug/useful info to http client.
// but doesn't have some of the extra sensitive info like env dump
// and host name or echo delay or header setting options.
func SafeDebugHandler(w http.ResponseWriter, r *http.Request) {
	fhttp.LogRequest(r, "Debug")
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
	buf.WriteString(fhttp.TLSInfo(r))
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
