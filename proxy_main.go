// Fortio TLS Reverse Proxy.
//
// (c) 2022 Laurent Demailly
// See LICENSE

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
	"fortio.org/fortio/version"
	"fortio.org/proxy/config"
)

func GetRoutes() []config.Route {
	routes := configs.Get().(*[]config.Route)
	return *routes
}

func setDestination(req *http.Request, url *url.URL) {
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
}

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

var (
	configs        = dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	email          = flag.String("email", "", "`Email` to attach to cert requests.")
	fullVersion    = flag.Bool("version", false, "Show full version info and exit.")
	certsFor       = dflag.DynStringSet(flag.CommandLine, "certs-domains", []string{}, "Coma seperated list of `domains` to get certs for")
	certsDirectory = flag.String("certs-directory", ".", "Directory `path` where to store the certs")
	port           = flag.String("https-port", ":443", "`port` to listen on for main reverse proxy and tls traffic")
	redirect       = flag.String("redirect-port", ":80", "`port` to listen on for redirection")
	configDir      = flag.String("config", "",
		"Config directory `path` to watch for changes of dynamic flags (empty for no watch)")
	httpPort = flag.String("http-port", "disabled", "`port` to listen on for non tls traffic (or 'disabled')")
	h2Target = flag.Bool("h2", false, "Whether destinations support h2c prior knowledge")
	acert    *autocert.Manager
)

func hostPolicy(ctx context.Context, host string) error {
	log.LogVf("cert host policy called for %q", host)
	allowed := certsFor.Get()
	if _, found := allowed[host]; found {
		return nil
	}
	return fmt.Errorf("acme/autocert: %q not in allowed list", host)
}

func printRoutes() {
	if !log.Log(log.Info) {
		return
	}
	log.Printf("Initial Routes:")
	for _, r := range GetRoutes() {
		log.Printf("host %q\t prefix %q\t -> %s", r.Host, r.Prefix, r.Destination.URL.String())
	}
}

func debugGetCert(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	log.LogVf("GetCert from %s for %q", hello.Conn.RemoteAddr().String(), hello.ServerName)
	return acert.GetCertificate(hello)
}

func usage(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "Fortio proxy %s usage:\n\t%s [flags]\nflags (some flags inherited from fortio but not used):\n",
		version.Short(),
		os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	flag.Parse()
	_, longV, fullV := version.FromBuildInfo()
	if len(flag.Args()) != 0 {
		usage("Only flags are expected")
	}
	if *fullVersion {
		fmt.Print(fullV)
		os.Exit(0)
	}
	if *configDir != "" {
		if _, err := configmap.Setup(flag.CommandLine, *configDir); err != nil {
			log.Critf("Unable to watch config/flag changes in %v: %v", *configDir, err)
		}
	}
	log.Printf("Fortio Proxy %s starting", longV)
	if *redirect != "disabled" {
		fhttp.RedirectToHTTPS(*redirect)
	}

	printRoutes()
	rp := httputil.ReverseProxy{Director: Director}

	// TODO: make h2c vs regular client more dynamic based on route config instead of all or nothing
	// (or maybe some day it will just ge the default behavior of the base http client)
	if *h2Target {
		rp.Transport = &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	}
	s := &http.Server{
		// TODO: make these timeouts configurable
		ReadTimeout:       6 * time.Second,
		WriteTimeout:      6 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		// The reverse proxy
		Handler: &rp,
	}
	if *httpPort != "disabled" {
		fhttp.HTTPServerWithHandler("http-reverse-proxy", *httpPort, &rp)
	}
	if *port == "disabled" {
		log.Infof("No TLS server port.")
		select {}
	}
	startTLSProxy(s)
}

func startTLSProxy(s *http.Server) {
	s.Addr = *port
	acert = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(*certsDirectory),
		Email:      *email,
	}
	tlsCfg := acert.TLSConfig()
	tlsCfg.GetCertificate = debugGetCert
	tlsCfg.MinVersion = tls.VersionTLS12
	s.TLSConfig = tlsCfg
	currentMap := certsFor.Get()
	currentDomains := make([]string, len(currentMap))
	i := 0
	for k := range currentMap {
		currentDomains[i] = k
		i++
	}
	log.Infof("Starting TLS on %s for %v", *port, currentDomains)
	err := s.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("ListendAndServeTLS(): %v", err)
	}
}
