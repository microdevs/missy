package main

import (
	"flag"
	"fmt"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = initCmd.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = initCmd.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])

		// init
		// pull up vault
		vaultContainerPorts := []apiv1.ContainerPort{
			{
				Name:          "http",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: 8200,
			},
		}

		vaultServicePorts := []apiv1.ServicePort{
			{
				Name:     "http",
				Protocol: apiv1.ProtocolTCP,
				Port:     80,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8200,
				},
			},
		}
		deploy("vault", "vault-server", "vault:0.9.0", vaultContainerPorts, vaultServicePorts, "missy", kubeconfig)
		// pull up some datastore

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
				Name:          "https-port",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: 8443,
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

		consulServicePorts := []apiv1.ContainerPort{}

		deploy("consul", "consul-server", "consul", consulContainerPorts, consulServicePorts, "missy", kubeconfig)
		// pull up missy-controller
	}
	os.Exit(0)
}

func deploy(name string, appName string, image string, containerPorts []apiv1.ContainerPort, servicePorts []apiv1.ServicePort, tier string, kubeconfig *string) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	deploymentsClient := clientset.AppsV1beta1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": name,
						"app":  appName,
						"tier": tier,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  name,
							Image: image,
							Ports: containerPorts,
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating Deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	fmt.Println("Creating Service ...")

	serviceClient := clientset.CoreV1().Services("default")

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":  appName,
				"tier": tier,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type:  apiv1.ServiceTypeClusterIP,
			Ports: servicePorts,
			Selector: map[string]string{
				"name": name,
			},
		},
	}

	_, err = serviceClient.Create(service)
	if err != nil {
		fmt.Errorf("failed to create service: %s", err)
		os.Exit(1)
	}
	fmt.Println("service created")
}

func int32Ptr(i int32) *int32 { return &i }
