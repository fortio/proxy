//go:build no_tailscale
// +build no_tailscale

package config

const HasTailscale = false

func IsTailscale(_ string) bool {
	return false
}

func Tailscale() CertificateProvider {
	return nil
}
