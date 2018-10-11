package missy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/microdevs/missy/log"
	"io/ioutil"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"os"
)

type CAConfig struct {
	PemBundle string `json:"pem_bundle"`
}


func CreateVaultConfig() {
	UploadCertificateToSecret("vault", "missy")
	vaultConfig := map[string]string{"vault.config": `{"listener":{"tcp":{"address":"0.0.0.0:8201","tls_cert_file":"/etc/vault-tls/vault.pem","tls_key_file":"/etc/vault-tls/vault.key"}},"backend": {"file": {"path": "/vault/file"}}, "default_lease_ttl": "168h", "max_lease_ttl": "720h"}`}
	Configmap("vault-config", "missy", vaultConfig)
}

func RunVault() {
	// pull up vault
	vaultContainerPorts := []apiv1.ContainerPort{
		{
			Name:          "https",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8200,
		},
	}

	vaultServicePorts := []apiv1.ServicePort{
		{
			Name:     "https",
			Protocol: apiv1.ProtocolTCP,
			Port:     443,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 8200,
			},
		},
	}

	 var envVars []apiv1.EnvVar

	 volumes := []apiv1.Volume{
		 {
			 Name: "vault-tls",
			 VolumeSource: apiv1.VolumeSource{
				 Secret: &apiv1.SecretVolumeSource{
					 SecretName: "vault-tls",
				 },
			 },
		 },
	 }

	 volumeMounts := []apiv1.VolumeMount{
	 	{
			 Name: "vault-tls",
			 MountPath: "/etc/vault-tls",
			 ReadOnly: true,
		 },
	 }

	 securityContext := &apiv1.SecurityContext{
		 Capabilities: &apiv1.Capabilities{
			 Add: []apiv1.Capability{
				 apiv1.Capability("IPC_LOCK"),
			 },
		 },
	 }

	Deploy("vault","missy", "vault-server", "vault:0.9.0", []string{"server"}, envVars, volumes, volumeMounts, vaultContainerPorts, vaultServicePorts, "missy", securityContext)
}

func InitVault() {
	api.NewClient(&api.Config{

	})
}

func ConfigureIntermediateRootCA() {

	caCertPem, err := ioutil.ReadFile(os.Getenv("HOME")+"/.missy/root-cert.pem")
	if err != nil {
		panic(fmt.Sprintf("cannot load root cert", err))
	}
	block, _ := pem.Decode(caCertPem)
	if block == nil {
		panic("cannot decode pem block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("unable to parse ca certificate", err))
	}

	sysCertPool, err := x509.SystemCertPool()
	if err != nil {
		panic("Cannot load system cert pool")
	}

	sysCertPool.AddCert(cert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs:      sysCertPool,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	httpClient := &http.Client{
		Transport: transport,
	}

	client, err := api.NewClient(
		&api.Config{
			Address:"https://vault:8200",
			HttpClient: httpClient,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("couldnt create vault client", err))
	}

	client.SetToken("4ba284b8-067c-3acf-73b1-ec82a10cda8b")

	ic, err := ioutil.ReadFile(os.Getenv("HOME")+"/.missy/intermediate-cert.pem")
	if err != nil {
		panic("Cannot load intermediate cert")
	}

	ik, err := ioutil.ReadFile(os.Getenv("HOME")+"/.missy/intermediate-key.pem")
	if err != nil {
		panic("Cannot load intermediate key")
	}

	pemBundle := string(ik)+string(ic)

	data := map[string]interface{}{
		"pem_bundle": pemBundle,
	}

	secret, error := client.Logical().Write("/pki/config/ca", data)
	if error != nil {
		log.Panicf("Error writing to vault %s", error)
	}
	log.Infof("%v", secret)

}
