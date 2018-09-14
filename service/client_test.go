package service

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fail()
	}
}

func TestRootCAs(t *testing.T) {
	// this is our test CA certificate
	caFileName := "test-fixtures/ca.crt"
	// set the env var make rootCAs() include it in the certpool
	os.Setenv("TLS_CAFILE", caFileName)
	// load the certificate as pem encoded []byte
	caPemEncodedCert, err := ioutil.ReadFile(caFileName)
	if err != nil {
		t.Fatal(err)
	}
	// get the system certpool
	certPool := rootCAs()
	// to be able to compare the subject we now parse our CA certificate
	var block *pem.Block
	block, _ = pem.Decode(caPemEncodedCert)
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	// now loop through the whole certpool's subjects and look out for our CA certificate
	for _, rawSubject := range certPool.Subjects() {
		if bytes.Equal(rawSubject, caCert.RawSubject) {
			// found it we can return
			return
		}
	}
	// if we dont find it we mark this test failed
	t.Log("custom cert not found in certpool")
	t.Fail()
}
