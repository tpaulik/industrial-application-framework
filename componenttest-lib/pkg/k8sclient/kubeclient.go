/*
Copyright 2020 Nokia
Licensed under the BSD 3-Clause License.
SPDX-License-Identifier: BSD-3-Clause
*/

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
