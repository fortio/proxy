package config

import "crypto/tls"

type CertGetter interface {
	GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
}
