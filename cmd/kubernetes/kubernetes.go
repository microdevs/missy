package kubernetes

import (
	"fmt"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	typed "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/microdevs/missy/log"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
	"time"
)

type ResourceClient interface {
	Create(name string, data map[string][]byte) (typed.ScaleInterface, error)
	Get(name string) (typed.ScaleInterface, error)
	Delete(name string) error
}

var kubeconfig *string

func InitializeBaseSystem(kcfg *string) {
	kubeconfig = kcfg

	// init
	// RunVault()
	// InitVault()
	// RunConsul()
}


func Deploy(name string, namespace string, appName string, image string, args []string, envVars []apiv1.EnvVar, volumes []apiv1.Volume, volumeMounts []apiv1.VolumeMount, containerPorts []apiv1.ContainerPort, servicePorts []apiv1.ServicePort, tier string, securityContext *apiv1.SecurityContext) {

	EnsureNamespace(namespace)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		panic(err)
	}



	deploymentsClient := clientset.AppsV1beta1().Deployments(namespace)

	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
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
							Args: args,
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

	serviceClient := clientset.CoreV1().Services(namespace)

	service := prepareService(name, appName, tier, servicePorts, apiv1.ServiceTypeClusterIP)

	err = serviceClient.Delete(service.Name, &metav1.DeleteOptions{
	})
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


func prepareService(name string, appName string, tier string, servicePorts []apiv1.ServicePort, serviceType apiv1.ServiceType) *apiv1.Service {

	s :=  &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: servicePorts,
			Selector: map[string]string{
				"app": appName,
			},
		},
	}

	switch serviceType {
	case apiv1.ClusterIPNone:
		s.Spec.ClusterIP = apiv1.ClusterIPNone
	default:
		s.Spec.Type = apiv1.ServiceTypeClusterIP
	}

	return s
}

func GetStorageClasses(kcfg *string) ([]v1beta1.StorageClass, error) {
	kubeconfig = kcfg
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	var classes []v1beta1.StorageClass

	storageClassList, err := clientset.StorageV1beta1().StorageClasses().List(metav1.ListOptions{})
	if err != nil {
		return classes, err
	}

	for _,sc := range storageClassList.Items {
		classes = append(classes, sc)
	}
	return classes, nil
}