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
	"os"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/dflag"
	"fortio.org/fortio/fhttp"
	"fortio.org/log"
	"fortio.org/proxy/config"
	"fortio.org/proxy/rp"
	"fortio.org/scli"
	"golang.org/x/crypto/acme/autocert"
)

const (
	disabled = "disabled"
)

var (
	email    = dflag.DynString(flag.CommandLine, "email", "", "`Email` to attach to cert requests.")
	certsFor = dflag.DynStringSet(flag.CommandLine, "certs-domains", []string{},
		"Coma separated list of `domains` to get certs for")
	certsDirectory = flag.String("certs-directory", ".", "Directory `path` where to store the certs")
	port           = flag.String("https-port", ":443", "`port` to listen on for main reverse proxy and tls traffic")
	redirect       = flag.String("redirect-port", ":80", "`port` to listen on for redirection")
	httpPort       = flag.String("http-port", disabled, "`port` to listen on for non tls traffic (or 'disabled')")
	autoTailscale  = flag.Bool("tailscale", false, "Automatically add tailscale hostname to the certificate list")
	timeout        = flag.Duration("timeout", 1*time.Minute,
		"Maximum duration for each request read/writes proxying (eg 1h or use 0 for no timeout)")
	acert     *autocert.Manager
	tailscale string
)

func hostPolicy(_ context.Context, host string) error {
	log.LogVf("cert host policy called for %q", host)
	if tailscale != "" && host == tailscale {
		return nil
	}
	allowed := certsFor.Get()
	if _, found := allowed[host]; found {
		return nil
	}
	return fmt.Errorf("acme/autocert: %q not in allowed list", host)
}

func debugGetCert(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// Note: hello.ServerName is already lowercase.
	isTailscale := config.IsTailscale(hello.ServerName)
	log.LogVf("GetCert from %s for %q (tailscale %t)",
		hello.Conn.RemoteAddr().String(), hello.ServerName, isTailscale)
	if isTailscale {
		if err := hostPolicy(context.Background(), hello.ServerName); err != nil {
			return nil, err
		}
		return config.Tailscale().GetCertificate(hello)
	}
	return acert.GetCertificate(hello)
}

func main() {
	cli.ProgramName = "Fortio proxy"
	scli.ServerMain()
	// Only turns on debug host if configured at launch,
	// can be turned off or changed later through dynamic flags but not turned on if starting off
	debugHost := rp.DebugHost.Get()
	if *redirect != disabled {
		var a net.Addr
		if debugHost != "" {
			// Special case for debug host, redirect to https but also serve debug on that host
			a = fhttp.HTTPServerWithHandler("https redirector + debug", *redirect, rp.DebugOnHostHandler(fhttp.RedirectToHTTPSHandler))
		} else {
			// Standard redirector without special debug host case
			a = fhttp.RedirectToHTTPS(*redirect)
		}
		if a == nil {
			os.Exit(1) // Error already logged
		}
	}
	if *autoTailscale {
		tailscale = config.TailscaleServerName()
		if tailscale == "" {
			os.Exit(1) // Error already logged
		}
		log.S(log.Info, "Will accept TLS requests and obtain certificate for tailscale", log.Any("server-name", tailscale))
	}
	// Main reverse proxy handler (with debug if configured)
	var hdlr http.Handler
	hdlr = rp.ReverseProxy()
	if debugHost != "" {
		log.Warnf("Running Debug echo handler for any request matching Host %q", debugHost)
		hdlr = rp.DebugOnHostHandler(hdlr.ServeHTTP) // that's the reverse proxy + debug handler
	}

	s := &http.Server{
		ReadTimeout:       *timeout,
		WriteTimeout:      *timeout,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 3 * time.Second, // reasonably small as the header are sent quickly for valid clients.
		// The reverse proxy (+debug if configured)
		Handler:  hdlr,
		ErrorLog: log.NewStdLogger("rp", log.Error),
	}

	log.Printf("Fortio Proxy %s started - hostid %q - tailscale capable: %t, on: %t", cli.LongVersion, rp.HostID.Get(), config.HasTailscale, *autoTailscale)

	if *httpPort != disabled {
		fhttp.HTTPServerWithHandler("http-reverse-proxy", *httpPort, hdlr)
	}

	if *port == disabled {
		log.Infof("No TLS server port.")
	} else {
		go startTLSProxy(s)
	}
	scli.UntilInterrupted()
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
