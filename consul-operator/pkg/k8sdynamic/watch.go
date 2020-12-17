// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

func WatchInformer(name string, namespace string, resourceVersion string, gvr schema.GroupVersionResource, eventHandle cache.ResourceEventHandler, stopper chan struct{}) {
	logger := log.WithName("WatchInformer").WithValues("resource", name)

	listOptionsFunc := dynamicinformer.TweakListOptionsFunc(func(options *v1.ListOptions) {
		if name != "" {
			options.FieldSelector = "metadata.name=" + name
		}

		if options.ResourceVersion == "0" && resourceVersion != "" {
			options.ResourceVersion = resourceVersion
		}
	})

	dynInformer := dynamicinformer.NewFilteredDynamicInformer(
		GetDynamicK8sClient(),
		gvr,
		namespace,
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		listOptionsFunc)

	informer := dynInformer.Informer()
	informer.AddEventHandler(eventHandle)

	informer.Run(stopper)

	logger.Info("resource watch has been stopped")
}
