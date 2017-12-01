package main

import (
	"flag"
	"os"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/api/core/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		vault(kubeconfig)
		// pull up some datastore
		// pull up missy-controller
	}
	os.Exit(0)
}

func vault(kubeconfig *string) {

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
			Name: "vault",
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "vault",
						"tier": "missy",
					},
				},
				Spec: apiv1.PodSpec{
					Containers:[]apiv1.Container{
						{
							Name: "vault",
							Image: "vault:0.9.0",
							Ports: []apiv1.ContainerPort{
								{
									Name: "http",
									Protocol: apiv1.ProtocolTCP,
									ContainerPort: 8200,
								},
							},
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
			Kind: "Service",
			APIVersion:"v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault",
			Labels: map[string]string{
				"app": "vault",
				"tier": "missy",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Protocol: apiv1.ProtocolTCP,
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: int32(8200),
					},
				},
			},
			Selector: map[string]string{
				"name": "vault",
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
