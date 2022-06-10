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
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"fortio.org/fortio/dflag"
	"fortio.org/fortio/dflag/configmap"
	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/log"
	"fortio.org/fortio/version"
	"fortio.org/proxy/rp"
)

var (
	email          = dflag.DynString(flag.CommandLine, "email", "", "`Email` to attach to cert requests.")
	certsFor       = dflag.DynStringSet(flag.CommandLine, "certs-domains", []string{}, "Coma seperated list of `domains` to get certs for")
	fullVersion    = flag.Bool("version", false, "Show full version info and exit.")
	certsDirectory = flag.String("certs-directory", ".", "Directory `path` where to store the certs")
	port           = flag.String("https-port", ":443", "`port` to listen on for main reverse proxy and tls traffic")
	redirect       = flag.String("redirect-port", ":80", "`port` to listen on for redirection")
	configDir      = flag.String("config", "",
		"Config directory `path` to watch for changes of dynamic flags (empty for no watch)")
	httpPort = flag.String("http-port", "disabled", "`port` to listen on for non tls traffic (or 'disabled')")
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

func debugGetCert(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	log.LogVf("GetCert from %s for %q", hello.Conn.RemoteAddr().String(), hello.ServerName)
	return acert.GetCertificate(hello)
}

func usage(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "Fortio proxy %s usage:\n\t%s [flags]\nflags (some flags inherited from fortio but not used):\n",
		version.Short(),
		os.Args[0])
	flag.PrintDefaults()
	if msg != "" {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

func main() {
	flag.CommandLine.Usage = func() { usage("") }
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

	rp.PrintRoutes()
	revp := rp.ReverseProxy()

	s := &http.Server{
		// TODO: make these timeouts configurable
		ReadTimeout:       6 * time.Second,
		WriteTimeout:      6 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		// The reverse proxy
		Handler: revp,
	}

	if *httpPort != "disabled" {
		fhttp.HTTPServerWithHandler("http-reverse-proxy", *httpPort, revp)
	}

	if *port == "disabled" {
		log.Infof("No TLS server port.")
		select {}
	}
	startTLSProxy(s)
}

func startTLSProxy(s *http.Server) {
	s.Addr = *port
	emailStr := strings.TrimSpace(email.Get())
	acert = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(*certsDirectory),
		Email:      emailStr,
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
	log.Infof("Starting TLS on %s for %v (%s) - certs directory %s", *port, currentDomains, acert.Email, *certsDirectory)
	err := s.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("ListendAndServeTLS(): %v", err)
	}
}
