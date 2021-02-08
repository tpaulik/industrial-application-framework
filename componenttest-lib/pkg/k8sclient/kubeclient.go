package k8sclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetK8sClient(c *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
