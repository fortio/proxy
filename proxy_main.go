package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"golang.org/x/crypto/acme/autocert"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
	"fortio.org/proxy/config"
)

func main() {
	configs := dflag.DynJSON(flag.CommandLine, "routes.json", &[]config.Route{}, "json list of `routes`")
	fullVersion := flag.Bool("version", false, "Show full version info")
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
	}
	if *configDir != "" {
		if _, err := configmap.Setup(flag.CommandLine, *configDir); err != nil {
			log.Critf("Unable to watch config/flag changes in %v: %v", *configDir, err)
		}
	}
	routes := configs.Get().(*[]config.Route)
	for i, r := range *routes {
		log.Infof("Route %d: %+v", i, r)
	}
	hostPolicy := func(ctx context.Context, host string) error {
		log.Infof("cert host policy called for %q", host)
		allowed := certsFor.Get()
		if _, found := allowed[host]; found {
			return nil
		}
		return fmt.Errorf("acme/autocert: only %v are is allowed", allowed)
	}

	m := http.NewServeMux()
	s := &http.Server{} // TODO: timeouts etc
	m.HandleFunc("/", fhttp.LogAndCall("debug", fhttp.DebugHandler))
	acert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(*certsDirectory),
	}
	debugGetCert := func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Infof("Called get cert with %+v", hello)
		return acert.GetCertificate(hello)
	}
	s.Addr = *port
	s.TLSConfig = &tls.Config{GetCertificate: debugGetCert}
	log.Infof("Starting TLS on %s", *port)
	err := s.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("ListendAndServeTLS(): %v", err)
	}
}
