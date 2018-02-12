package kubernetes

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RunConsul() {
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
			Name:     "ui-port",
			Protocol: apiv1.ProtocolTCP,
			Port:     8500,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 8500,
			},
		},
	}

	deploy("consul", "consul-server", "consul", []apiv1.EnvVar{}, []apiv1.Volume{}, []apiv1.VolumeMount{}, consulContainerPorts, consulServicePorts, "missy", nil)
}
