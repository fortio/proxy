//go:build !no_tailscale
// +build !no_tailscale

package config

// build constraints

import (
	"context"
	"strings"

	"fortio.org/log"
	"tailscale.com/client/tailscale"
)

const HasTailscale = true

// Suffix for server names which will use the tailscale client instead of the autocert client.
// Not expected to be changed but just in case.
var TailscaleSuffix = ".ts.net"

// IsTailscale returns true if the server name ends with the TailscaleSuffix.
// Note the check is expecting lowercase serverName which is what hello.ServerName already is.
func IsTailscale(serverName string) bool {
	return strings.HasSuffix(serverName, TailscaleSuffix)
}

var tcli = &tailscale.LocalClient{}

func Tailscale() CertificateProvider {
	return tcli
}

func TailscaleServerName() string {
	status, err := tcli.StatusWithoutPeers(context.Background())
	if err != nil {
		log.Critf("Error getting tailscale status: %v", err)
		return ""
	}
	// Remove the trailing dot as it's not there in ServerName.
	return strings.TrimSuffix(status.Self.DNSName, ".")
}
