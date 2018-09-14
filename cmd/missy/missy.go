package missy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/microdevs/missy/log"
	"math/big"
	"os"
	"time"
)

var configDir = func() string {
	home := os.Getenv("HOME")
	if home == "" {
		panic("No $HOME environment variable specified")
	}
	path := home + "/.missy"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0700)
		if err != nil {
			panic("Error creating missy dir")
		}
	}
	return path
}()

func CreateRootCA() {

	rootCATargetFile := configDir+"/root-cert.pem"

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Missy"},
			CommonName: "Missy SubSystem",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: true,

	}

	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(rootCATargetFile)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Info("written root-cert.pem")

	keyOut, err := os.OpenFile(configDir+"/root-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print("failed to open key.pem for writing:", err)
		return
	}
	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()
	log.Info("written root-key.pem")

}

func CreateCertFromCA(name, rootCertFile, rootKeyFile, o, ou, cn, country, province string, isCA bool) {
	// Load CA
	root, err := tls.LoadX509KeyPair(configDir+"/"+rootCertFile, configDir+"/"+rootKeyFile)
	if err != nil {
		panic(err)
	}
	rootCert, err := x509.ParseCertificate(root.Certificate[0])
	if err != nil {
		panic(err)
	}

	intermediateCAKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{o},
			OrganizationalUnit: []string{ou},
			CommonName: cn,
			Country: []string{country},
			Province: []string{province},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	imcert, err := x509.CreateCertificate(rand.Reader, &template, rootCert, publicKey(intermediateCAKey), root.PrivateKey)
	if err != nil {
		panic("unable to issue intermediate CA cert")
	}

	certFileName := fmt.Sprintf("%s-cert.pem", name)
	certFilePath := configDir+"/"+certFileName
	certOut, err := os.Create(certFilePath)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", certFileName, err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: imcert})
	certOut.Close()
	log.Infof("written %s", certFileName)

	keyFileName := fmt.Sprintf("%s-key.pem", name)
	keyOut, err := os.OpenFile(configDir+"/"+keyFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Infof("failed to open %s for writing:", keyFileName, err)
		return
	}
	pem.Encode(keyOut, pemBlockForKey(intermediateCAKey))
	keyOut.Close()
	log.Infof("written %s", keyFileName)

}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}