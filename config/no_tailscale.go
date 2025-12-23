//go:build no_tailscale

package config

const HasTailscale = false

func IsTailscale(_ string) bool {
	return false
}

func Tailscale() CertificateProvider {
	return nil
}

func TailscaleServerName() string {
	panic("Binary built without tailscale support, rebuild without -tags no_tailscale")
}
