// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	"context"
	regexp2 "regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const ServiceKind = "Service"

func (k K8sDynClient) ApplyConcatenatedResources(resourcesStr string, namespace string) ([]ResourceDescriptor, error) {
	var resourceDescriptors []ResourceDescriptor

	yamlResources := splitToIndividualResources(resourcesStr)
	for _, yamlRes := range yamlResources {
		yamlRes = removeCommentedParts(yamlRes)
		yamlRes = strings.Trim(yamlRes, "\n")

		if yamlRes == "" {
			continue
		}
		log.V(1).Info("Resource to apply", "content", yamlRes)

		resourceDescriptor, err := k.ApplyYamlResource(yamlRes, namespace)
		if err != nil {
			return resourceDescriptors, err
		}
		resourceDescriptors = append(resourceDescriptors, resourceDescriptor)
	}
	return resourceDescriptors, nil
}

func removeCommentedParts(s string) string {
	logger := log.WithName("removedCommentedParts")
	regexp, err := regexp2.Compile("(?m)#.*$")
	if err != nil {
		logger.Error(err, "Failed to compile regexp")
		return s
	}
	return regexp.ReplaceAllString(s, "")
}

func (k K8sDynClient) ApplyYamlResource(resourceStr string, namespace string) (ResourceDescriptor, error) {
	object, err := yamlToUnstructured(resourceStr)
	if err != nil {
		return ResourceDescriptor{}, err
	}

	resourceDescriptor, err := k.applyResource(&object, namespace)
	if err != nil {
		return ResourceDescriptor{}, err
	}

	return resourceDescriptor, nil
}

func (k *K8sDynClient) applyResource(object *unstructured.Unstructured, namespace string) (ResourceDescriptor, error) {
	logger := log.WithName("applyResource")
	gvk := object.GroupVersionKind()

	apiResource, err := k.getAPIResourceByGvk(gvk)
	if err != nil {
		return ResourceDescriptor{}, errors.Wrap(err, "failed to find the resource by gvk")
	}

	gvr := GroupVersionResource{Version: gvk.Version, Group: gvk.Group, Resource: apiResource.Name}
	logger.Info("GVR of the app specific CR", "value", gvr)

	resourceDescriptor := ResourceDescriptor{
		Name: object.GetName(),
		Gvr:  gvr,
	}

	var k8sResource dynamic.ResourceInterface
	if apiResource.Namespaced {
		k8sResource = k.dynClient.Resource(gvr.GetGvr()).Namespace(namespace)
		resourceDescriptor.Namespace = namespace
	} else {
		k8sResource = k.dynClient.Resource(gvr.GetGvr())
	}

	actVer, err := k8sResource.Get(context.TODO(), object.GetName(), metav1.GetOptions{})
	if err != nil {
		logger.Info("resource doesn't exist, create it")
		_, err = k8sResource.Create(context.TODO(), object, metav1.CreateOptions{})
	} else {
		logger.Info("resource already exist, update it")
		object.SetResourceVersion(actVer.GetResourceVersion())
		if gvk.Kind == ServiceKind {
			outBytes, err2 := runtime.Encode(unstructured.UnstructuredJSONScheme, object)
			if err2 != nil {
				return ResourceDescriptor{}, err
			}
			_, err = k8sResource.Patch(context.TODO(), object.GetName(), types.MergePatchType, outBytes, metav1.PatchOptions{})
		} else {
			_, err = k8sResource.Update(context.TODO(), object, metav1.UpdateOptions{})
		}
	}

	if err != nil {
		return resourceDescriptor, errors.Wrap(err, "failed to apply the given resource")
	}

	return resourceDescriptor, nil
}

func (k K8sDynClient) getAPIResourceByGvk(gvk schema.GroupVersionKind) (metav1.APIResource, error) {
	if gvk.Version == "" || gvk.Kind == "" {
		return metav1.APIResource{}, errors.New("empty input parameters")
	}

	groupVersion := ""
	if gvk.Group == "" {
		groupVersion = gvk.Version
	} else {
		groupVersion = gvk.Group + "/" + gvk.Version
	}

	resList, err := k.generalClient.Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return metav1.APIResource{}, err
	}

	for _, res := range resList.APIResources {
		if res.Kind == gvk.Kind {
			return res, nil
		}
	}

	return metav1.APIResource{}, errors.New("not found")
}
