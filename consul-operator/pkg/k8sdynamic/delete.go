// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sDynClient) DeleteResources(appliedResources []ResourceDescriptor) error {
	logger := log.WithName("DeleteResources")

	for _, appliedResource := range appliedResources {
		logger.Info("resourceDescriptor", "value", appliedResource)
		_, err := k.dynClient.Resource(appliedResource.Gvr.GetGvr()).Namespace(appliedResource.Namespace).Get(appliedResource.Name, metav1.GetOptions{})
		if err != nil {
			logger.Info("resource doesn't exist")
		} else {
			deletePolicy := metav1.DeletePropagationBackground
			gracePeriodSeconds := int64(0)
			deleteOptions := &metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriodSeconds,
				PropagationPolicy:  &deletePolicy,
			}
			if err := k.dynClient.Resource(appliedResource.Gvr.GetGvr()).Namespace(appliedResource.Namespace).Delete(appliedResource.Name, deleteOptions); err != nil {
				return err
			}
		}
	}

	return nil
}
