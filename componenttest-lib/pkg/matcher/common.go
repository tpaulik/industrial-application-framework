/*
Copyright 2020 Nokia
Licensed under the BSD 3-Clause License.
SPDX-License-Identifier: BSD-3-Clause
*/

package matcher

import (
	"context"
	"github.com/pkg/errors"
	"github.com/nokia/industrial-application-framework/componenttest-lib/pkg/env"
	"github.com/nokia/industrial-application-framework/componenttest-lib/pkg/k8sclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type K8sResourceId struct {
	Name      string
	Namespace string
	ParamPath []string
	Gvk       schema.GroupVersionKind
}

func GetCurrentStateOfResource(apiResource v1.APIResource, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
	k8sClient := k8sclient.GetDynamicK8sClient(env.Cfg)
	var k8sResource dynamic.ResourceInterface
	if apiResource.Namespaced {
		k8sResource = k8sClient.Resource(gvr).Namespace(namespace)
	} else {
		k8sResource = k8sClient.Resource(gvr)
	}
	return k8sResource.Get(context.TODO(),name, v1.GetOptions{})
}

func GetGvrAndAPIResources(gvk schema.GroupVersionKind) (schema.GroupVersionResource, v1.APIResource, error) {
	apiResource, err := GetAPIResourceByGvk(gvk)
	if err != nil {
		return schema.GroupVersionResource{}, v1.APIResource{}, errors.Wrap(err, "failed to get the apiResource of the given resource")
	}
	gvr := schema.GroupVersionResource{Version: gvk.Version, Group: gvk.Group, Resource: apiResource.Name}
	return gvr, apiResource, nil
}
