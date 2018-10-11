package kubernetes

import (
	"errors"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)
// SecretClient holds the active client to the Kubernetes Cluster
type SecretClient struct {
	client v1.SecretInterface
	namespace string
}

// NewSecretsClient will create a new instance with an active client to the Cluster
func NewSecretsClient(namespace string) *SecretClient {
	sc := SecretClient{}
	sc.namespace = namespace
	sc.client = ClientSet().CoreV1().Secrets(namespace)
	return &sc
}

// Create wraps around the original Kubernetes client func and makes sure an existing secret is deleted first
func (sc *SecretClient) Create(name string, data map[string][]byte) (*apiv1.Secret, error) {
	err := EnsureNamespace(sc.namespace)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Create Secret: Error during EnsureNamespace: %s",err))
	}
	//if exists delete
	if _,err := sc.Get(name); err == nil {
		err = sc.Delete(name)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Create Secret: Error deleting existing secret before creation: %s", err))
		}
	}
	// build the secret
	secret := &apiv1.Secret{
		Data: data,
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: sc.namespace,
		},
	}
	// actually create it
	return sc.client.Create(secret)
}

// Delete wraps around the original Kubernetes client func for easier handling
func (sc *SecretClient) Delete(name string) error {
	return sc.client.Delete(name, &metav1.DeleteOptions{})
}

// Get wraps around the original Kubernetes client func for easier handling
func (sc *SecretClient) Get(name string) (*apiv1.Secret, error) {
	return sc.client.Get(name, metav1.GetOptions{})
}