package kubernetes

import (
	"fmt"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureNamespace checks if a namespace exists and creates the namespace if it not exists
func EnsureNamespace(namespace string) error {
	_, err := ClientSet().CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err == nil {
		// if the namespace exists return silently doing nothing
		return nil
	}
	// create namespace spec
	nsSpec := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	// create the namespace
	_, err = ClientSet().CoreV1().Namespaces().Create(nsSpec)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while creating namespace %s: %s",namespace, err))
	}
	return nil
}
