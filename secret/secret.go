package secret

import (
	"crypto/x509"
	"github.com/microdevs/missy/service"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

func NewCertificate(cn string, expires *time.Time) (*x509.Certificate, error) {
	_, err := api.NewClient(&api.Config{})
	if err != nil {
		return nil, err
	}

}
