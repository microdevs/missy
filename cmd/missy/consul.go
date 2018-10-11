package missy

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/microdevs/missy/cmd/kubernetes"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1beta1"
	"log"

	"k8s.io/apimachinery/pkg/util/intstr"
)

type ConsulCompanion struct {

}


func (c *ConsulCompanion) Configure() error {
	kubernetes.UploadCertificateToSecret("consul", "missy")

	key := make([]byte, 16)
	n, err := rand.Reader.Read(key)
	if err != nil {
		log.Fatalf("Error reading random data: %s", err)
	}
	if n != 16 {
		log.Fatalf("Couldn't read enough entropy. Generate more entropy!")
	}

	keybytes := []byte(base64.StdEncoding.EncodeToString(key))

	keydata := map[string][]byte{
		"gossip-encryption-key": keybytes,
	}

	kubernetes.Secret("consul", "missy", keydata)
}


func (c *ConsulCompanion) Run(storageClass v1beta1.StorageClass) error {
	consulContainerPorts := []apiv1.ContainerPort{
		{
			Name:          "ui-port",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8500,
		},
		{
			Name:          "alt-port",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8400,
		},
		{
			Name:          "https-port",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8443,
		},
		{
			Name:          "http-port",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8080,
		},
		{
			Name:          "udp-port",
			Protocol:      apiv1.ProtocolUDP,
			ContainerPort: 53,
		},
		{
			Name:          "serflan",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8301,
		},
		{
			Name:          "serfwan",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8302,
		},
		{
			Name:          "consuldns",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8600,
		},
		{
			Name:          "server",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: 8300,
		},
	}

	consulServicePorts := []apiv1.ServicePort{
		{
			Name: "http",
			Port: 8500,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8500),
		},
		{
			Name: "https",
			Port: 8443,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8443),
		},
		{
			Name: "rpc",
			Port: 8400,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8400),
		},
		{
			Name: "serflan-tcp",
			Port: 8301,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8301),
		},
		{
			Name: "serflan-udp",
			Port: 8301,
			Protocol: apiv1.ProtocolUDP,
			TargetPort: intstr.FromInt(8301),
		},
		{
			Name: "serfwan-tcp",
			Port: 8302,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8302),
		},
		{
			Name: "serfwan-udp",
			Port: 8302,
			Protocol: apiv1.ProtocolUDP,
			TargetPort: intstr.FromInt(8302),
		},
		{
			Name: "server",
			Port: 8300,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8300),
		},
		{
			Name: "consuldns",
			Port: 8600,
			Protocol: apiv1.ProtocolTCP,
			TargetPort: intstr.FromInt(8600),
		},
	}


	/*volumes := []apiv1.Volume{
		{
			Name: "consul-tls",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: "consul-tls",
				},
			},
		},
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name: "consul-tls",
			MountPath: "/etc/consul-tls",
			ReadOnly: true,
		},
	}*/

	args := []string{
		"agent",
		"-advertise=$(POD_IP)",
		"-bind=0.0.0.0",
		"-bootstrap-expect=3",
		"-retry-join=consul-0.consul.$(NAMESPACE).svc.cluster.local",
		"-retry-join=consul-1.consul.$(NAMESPACE).svc.cluster.local",
		"-retry-join=consul-2.consul.$(NAMESPACE).svc.cluster.local",
		"-client=0.0.0.0",
		"-config-file=/etc/consul/config/server.json",
		"-datacenter=dc1",
		"-data-dir=/consul/data",
		"-domain=cluster.local",
		"-encrypt=$(GOSSIP_ENCRYPTION_KEY)",
		"-server",
		"-ui",
		"-disable-host-node-id",
		"-disable-keyring-file",
	}

	DeployStatefulSet("missy", "consul", args, consulContainerPorts, consulServicePorts, storageClass)
}
