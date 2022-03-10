// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var Config *rest.Config

func GetDynamicK8sClient() dynamic.Interface {
	if Config == nil {
		var err error
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err := dynamic.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
