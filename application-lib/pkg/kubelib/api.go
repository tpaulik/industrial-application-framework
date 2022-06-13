// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package kubelib

import (
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
