// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package kubelib

import (
	appsv1 "k8s.io/api/apps/v1"
	k8v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//GetKubeAPI returns the clientset for the container running inside kubernetes.
//The container must have the RBAC roles in place to use kubernetes api

var Config *rest.Config

func GetKubeAPI() *kubernetes.Clientset {
	// creates the in-cluster config
	if Config == nil {
		var err error
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(Config)
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
func CreateDeploymentConfig(name string) *appsv1.Deployment {
	dep := &appsv1.Deployment{}
	dep.Kind = "Deployment"
	dep.APIVersion = "apps/v1"
	dep.Name = name
	var revLimit int32
	dep.Spec.RevisionHistoryLimit = &revLimit
	var i int32 = 1
	dep.Spec.Replicas = &i
	return dep
}
