package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// KeyPair holds a certificate together with its private key and optional CA.
type KeyPair struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key" yaml:"key"`
	CA   string `json:"ca" yaml:"ca"`
}

// X509Certificate is a json/yaml marshallable/unmarshallable type wrapper for x509.Certificate.
type X509Certificate struct {
	*x509.Certificate
}

// ParseX509Certificate decodes the given PEM encoded string and parses it into an X509Certificate.
func ParseX509Certificate(certStr string) (*X509Certificate, error) {
	block, _ := pem.Decode([]byte(certStr))
	if block == nil {
		return nil, fmt.Errorf("Failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &X509Certificate{Certificate: cert}, nil
}
