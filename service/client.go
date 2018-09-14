package service

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/microdevs/missy/log"
)

// NewClient returns a new http.Client with custom CA Cert if given through TLS_CACERT
func NewClient() *http.Client {
	config := &tls.Config{
		RootCAs: rootCAs(),
	}
	tr := &http.Transport{TLSClientConfig: config}
	return &http.Client{Transport: tr}
}

// rootCAs returns the system CA Pool and inserts a custom CA file if given
func rootCAs() *x509.CertPool {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if rootCAs == nil || err != nil {
		rootCAs = x509.NewCertPool()
	}

	caFile := strings.Trim(os.Getenv("TLS_CAFILE"), " ")
	// do some checks on the CA file
	if caFile != "" {
		if _, err := os.OpenFile(caFile, os.O_RDONLY, 0600); os.IsNotExist(err) || os.IsPermission(err) {
			log.Warnf("TLS CA File is not accessible: %v, you may get TLS validation errors", err)
			return rootCAs
		}
		// Read in the cert file
		certs, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Warnf("Failed to read %q CA cert file: %v, you may get TLS validation errors", caFile, err)
			return rootCAs
		}
		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only, you may get TLS validation errors")
			return rootCAs
		}
		// return system pool + our custom CA
		return rootCAs
	}
	log.Warn("TLS CA File was not set, you may get TLS validation errors")
	return rootCAs
}
