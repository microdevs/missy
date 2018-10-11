package kubernetes

import (
	"github.com/microdevs/missy/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config reads kubeconfig from file and returns a pointer to rest.Config
func Config() *rest.Config {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		// todo: handle error better and return a proper info what happened
		log.Fatal(err)
	}

	return config
}
