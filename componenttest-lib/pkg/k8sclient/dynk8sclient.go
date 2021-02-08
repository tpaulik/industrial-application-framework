package k8sclient

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func GetDynamicK8sClient(c *rest.Config) dynamic.Interface {
	clientset, err := dynamic.NewForConfig(c)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
