package kubernetes

import (
	"fmt"
	"github.com/microdevs/missy/log"
	"io/ioutil"
	"os"

)

// int32Ptr returns a pointer from an int
func int32Ptr(i int32) *int32 {
	return &i
}

// UploadCertificateToSecret creates a secret "name-tls" with root-ca cert, server cert and server key
// that can be used by the service to enable tls encryption
func UploadCertificateToSecret(name string, namespace string) {

	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/.missy/%s-cert.pem", os.Getenv("HOME"), name))
	if err != nil {
		log.Panicf("failed to load %s cert: %s", name, err)
	}

	key, err := ioutil.ReadFile(fmt.Sprintf("%s/.missy/%s-key.pem", os.Getenv("HOME"), name))
	if err != nil {
		log.Panicf("failed to load %s key: %s", name, err)
	}

	caCert, err := ioutil.ReadFile(os.Getenv("HOME")+"/.missy/root-cert.pem")
	if err != nil {
		log.Panicf("failed to load root cert: %s", err)
	}


	data := map[string][]byte{
		name+".pem": cert,
		name+".key": key,
		"ca.pem": caCert,
	}

	Secret(name+"-tls", namespace, data)
}