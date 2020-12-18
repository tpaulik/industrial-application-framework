// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

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
