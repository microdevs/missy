package certificates

import (
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
)

// Pool returns the system CA Pool and inserts a custom CA if given.
func Pool(ca []byte) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if pool == nil || err != nil {
		pool = x509.NewCertPool()
	}
	if ca != nil {
		if ok := pool.AppendCertsFromPEM(ca); !ok {
			return pool, errors.New("couldn't append certs from pem")
		}
	}
	return pool, nil
}

// PoolFromFile returns system CA Pool and inserts a custom CA file if given.
func PoolFromFile(filename string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if pool == nil || err != nil {
		pool = x509.NewCertPool()
	}
	ca, err := ioutil.ReadFile(filename)
	if err != nil {
		return pool, errors.Wrapf(err, "couldn't open file with name='%s'", filename)
	}
	return Pool(ca)
}
