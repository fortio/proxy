package config

import "strings"

// Suffix for server names which will use the tailscale client instead of the autocert client.
// Not expected to be changed but just in case.
var TailscaleSuffix = ".ts.net"

// IsTailscale returns true if the server name ends with the TailscaleSuffix.
// Note the check is expecting lowercase serverName which is what hello.ServerName already is.
func IsTailscale(serverName string) bool {
	return strings.HasSuffix(serverName, TailscaleSuffix)
}
