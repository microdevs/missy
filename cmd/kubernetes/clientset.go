package kubernetes

import (
	"k8s.io/client-go/kubernetes"
)

// ClientSet returns a clientset with current config
func ClientSet() *kubernetes.Clientset {

	config := Config()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}