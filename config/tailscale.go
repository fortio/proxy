//go:build !no_tailscale
// +build !no_tailscale

package config

// build constraints

import (
	"strings"
	"sync"

	"tailscale.com/client/tailscale"
)

// Suffix for server names which will use the tailscale client instead of the autocert client.
// Not expected to be changed but just in case.
var TailscaleSuffix = ".ts.net"

// IsTailscale returns true if the server name ends with the TailscaleSuffix.
// Note the check is expecting lowercase serverName which is what hello.ServerName already is.
func IsTailscale(serverName string) bool {
	return strings.HasSuffix(serverName, TailscaleSuffix)
}

var (
	tcert     *tailscale.LocalClient
	tcertOnce sync.Once
)

func Tailscale() CertGetter {
	tcertOnce.Do(func() {
		tcert = &tailscale.LocalClient{}
	})
	return tcert
}
