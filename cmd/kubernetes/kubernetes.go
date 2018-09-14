package kubernetes

import (
	"fmt"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
	"github.com/microdevs/missy/log"
	"time"
)

var kubeconfig *string

func InitializeBaseSystem(kcfg *string) {
	kubeconfig = kcfg

	// init
	// RunVault()
	// InitVault()
	// RunConsul()
}

func deploy(name string, appName string, image string, envVars []apiv1.EnvVar, volumes []apiv1.Volume, volumeMounts []apiv1.VolumeMount, containerPorts []apiv1.ContainerPort, servicePorts []apiv1.ServicePort, tier string, securityContext *apiv1.SecurityContext) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := k8s.NewForConfig(config)
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
							Env: envVars,
							VolumeMounts: volumeMounts,
							SecurityContext: securityContext,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	err = deploymentsClient.Delete(deployment.Name, &metav1.DeleteOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Panicf("Error deleting existing service: %s", err)
		}
	} else {
		for {
			_, err := deploymentsClient.Get(deployment.Name, metav1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					break
				}
				log.Panicf("Error while waiting for deleting deployment: %s", err)
			}
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
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

	err = serviceClient.Delete(service.Name, &metav1.DeleteOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Panicf("Error deleting existing service: %s", err)
		}
	} else {
		for {
			_, err := serviceClient.Get(service.Name, metav1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					break
				}
				log.Panicf("Error while waiting for deletion of service: %s", err)
			}
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
	}


	_, err = serviceClient.Create(service)
	if err != nil {
		fmt.Errorf("failed to create service: %s", err)
		os.Exit(1)
	}
	fmt.Println("service created")
}

func int32Ptr(i int32) *int32 { return &i }
