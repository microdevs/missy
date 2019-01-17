package client

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/microdevs/missy/certificates"
	"github.com/pkg/errors"
)

type Config struct {
	CAFile  string        `env:"HTTP_CLIENT_CA_FILE"`
	Timeout time.Duration `env:"HTTP_CLIENT_TIMEOUT" envDefault:"600s"`
}

// New returns a new http.Client with custom CA Cert if given through TLS_CACERT.
// If rootCAs is empty, then default system ca will be used.
func New(c Config) (*http.Client, error) {
	if c.Timeout == 0 {
		c.Timeout = time.Minute * 10
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Errorf("getting system cert pool err: %s", err)
	}
	if c.CAFile != "" {
		cp, err := certificates.PoolFromFile(c.CAFile)
		if err != nil {
			return nil, errors.Errorf("getting cert pool from file ('%s') err: %s", c.CAFile, err)
		}
		rootCAs = cp
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
		Timeout: c.Timeout,
	}, nil
}
