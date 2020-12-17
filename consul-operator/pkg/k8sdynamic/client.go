package k8sdynamic

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func GetDynamicK8sClient() dynamic.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
