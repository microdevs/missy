package kubernetes

import (
	"errors"
	"fmt"
	"github.com/microdevs/missy/log"
	apiv1  "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
)

// SecretClient holds the active client to the Kubernetes Cluster
type ConfigMapClient struct {
	client v1.ConfigMapInterface
	namespace string
}

// NewSecretsClient will create a new instance with an active client to the Cluster
func NewConfigMapClient(namespace string) *ConfigMapClient {
	cmc := ConfigMapClient{}
	cmc.namespace = namespace
	cmc.client = ClientSet().CoreV1().ConfigMaps(namespace)
	return &cmc
}

func (cmc *ConfigMapClient) Create(name string, data map[string]string) (*apiv1.ConfigMap, error) {
	err := EnsureNamespace(cmc.namespace)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Create Configmap error during ensure namespace: %s", err))
	}

	// delete the existing one first
	err = cmc.client.Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Errorf("Error deleting existing config: %s", err)
			return nil, errors.New(fmt.Sprintf("Error during create Configmap: %s", err))
		}
	}
	// create the spec
	configMapSpec := &apiv1.ConfigMap{
		Data: data,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	config, err := cmc.client.Create(configMapSpec)
	if err != nil {
		log.Errorf("Error creating config: %s", err)
		return nil, errors.New(fmt.Sprintf("Error during create Configmap: %s", err))
	}

	return config, nil
}

func (cmc *ConfigMapClient) Delete(name string) error {
	return cmc.client.Delete(name, &metav1.DeleteOptions{})
}

func (cmc *ConfigMapClient) Get(name string) (*apiv1.ConfigMap, error) {
	return cmc.client.Get(name, metav1.GetOptions{})
}