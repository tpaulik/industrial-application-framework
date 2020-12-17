package kubelib

import (
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	k8v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//GetKubeAPI returns the clientset for the container running inside kubernetes.
//The container must have the RBAC roles in place to use kubernetes api
func GetKubeAPI() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

//CreateContainer creates a default container.
func CreateContainer(name string, image string) *k8v1.Container {
	cont := &k8v1.Container{}
	cont.Name = name
	cont.Image = image
	cont.ImagePullPolicy = "IfNotPresent"
	return cont
}

//CreateConfigMap creates a configmap that can be submitted to kubernetes
func CreateConfigMap(name string, dataKey string, data string) *k8v1.ConfigMap {
	c := &k8v1.ConfigMap{}
	c.Kind = "ConfigMap"
	c.APIVersion = "v1"
	c.Name = name
	c.Data = make(map[string]string)
	c.Data[dataKey] = data
	return c
}

//CreateDeployment
func CreateDeploymentConfig(name string) *appsv1beta2.Deployment {
	dep := &appsv1beta2.Deployment{}
	dep.Kind = "Deployment"
	dep.APIVersion = "apps/v1beta2"
	dep.Name = name
	var revLimit int32
	dep.Spec.RevisionHistoryLimit = &revLimit
	var i int32 = 1
	dep.Spec.Replicas = &i
	return dep
}
