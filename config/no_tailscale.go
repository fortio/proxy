//go:build no_tailscale
// +build no_tailscale

package config

func IsTailscale(_ string) bool {
	return false
}

func Tailscale() CertGetter {
	return nil
}
