// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type K8sDynClient struct {
	dynClient     dynamic.Interface
	generalClient *kubernetes.Clientset
}

type ResourceDescriptor struct {
	Name      string
	Namespace string
	Gvr       schema.GroupVersionResource
}

var log = logf.Log.WithName("k8sdynamic")

func New(genClient *kubernetes.Clientset) K8sDynClient {
	return K8sDynClient{
		dynClient:     GetDynamicK8sClient(),
		generalClient: genClient,
	}
}
