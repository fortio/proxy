package config

import "crypto/tls"

// CertificateProvider interface is used to hide the tailscale dependency when not using it,
// could also be use as a generic mapping between server name and cert provider
// (while right now it's either autocert or tailscale).
type CertificateProvider interface {
	GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
}
