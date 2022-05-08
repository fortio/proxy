package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/log"
	"fortio.org/proxy/config"
)

var (
	configs = dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
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
	log.LogVf("Directing %v", req)
	for _, route := range routes {
		log.LogVf("Evaluating %v", routes)
		if req.Host == route.Host || route.Host == "*" {
			setDestination(req, &route.Destination.URL)
		}
	}
}

func main() {
	email := flag.String("email", "", "`Email` to attach to cert requests.")
	fullVersion := flag.Bool("version", false, "Show full version info and exit.")
	certsFor := dflag.DynStringSet(flag.CommandLine, "certs-domains", []string{}, "Coma seperated list of `domains` to get certs for")
	certsDirectory := flag.String("certs-directory", ".", "Directory `path` where to store the certs")
	port := flag.String("port", ":443", "`port` to listen on")
	configDir := flag.String("config", "",
		"Config directory `path` to watch for changes of dynamic flags (empty for no watch)")
	flag.Parse()
	binfo, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatalf("Unexpected to be unable to get BuildInfo")
	}
	log.Infof("Fortio Proxy %s (%s %s %s) starting", binfo.Main.Version, binfo.GoVersion, runtime.GOARCH, runtime.GOOS)
	if *fullVersion {
		log.Infof("Buildinfo: %s", binfo.String())
		os.Exit(0)
	}
	if *configDir != "" {
		if _, err := configmap.Setup(flag.CommandLine, *configDir); err != nil {
			log.Critf("Unable to watch config/flag changes in %v: %v", *configDir, err)
		}
	}
	hostPolicy := func(ctx context.Context, host string) error {
		log.Infof("cert host policy called for %q", host)
		allowed := certsFor.Get()
		if _, found := allowed[host]; found {
			return nil
		}
		return fmt.Errorf("acme/autocert: only %v are is allowed", allowed)
	}
	log.Infof("Initial Routes:")
	for _, r := range GetRoutes() {
		log.Infof("%q -> %s", r.Host, r.Destination.URL.String())
	}
	rp := httputil.ReverseProxy{Director: Director}
	s := &http.Server{
		// TODO: make these timeouts configurable
		ReadTimeout:       6 * time.Second,
		WriteTimeout:      6 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		// The reverse proxy
		Handler: &rp,
	}
	acert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(*certsDirectory),
		Email:      *email,
	}
	debugGetCert := func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Infof("GetCert from %s for %q", hello.Conn.RemoteAddr().String(), hello.ServerName)
		return acert.GetCertificate(hello)
	}
	s.Addr = *port
	tlsCfg := acert.TLSConfig()
	tlsCfg.GetCertificate = debugGetCert
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