//go:build !no_tailscale
// +build !no_tailscale

package config_test

import (
	"testing"

	"fortio.org/proxy/config"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		serverName string
		match      bool
	}{
		{"foo.bar.ts.net", true},  // 0
		{"www.google.com", false}, // 1
	}
	for i, tst := range tests {
		res := config.IsTailscale(tst.serverName)
		if res != tst.match {
			t.Errorf("Mismatch %d expected %v for %s", i, tst.match, tst.serverName)
		}
	}
}
