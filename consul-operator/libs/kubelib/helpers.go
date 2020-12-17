// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package kubelib

import (
	k8v1 "k8s.io/api/core/v1"
)

//AddPullSecret adds a new pullsecret to the Pod
func AddPullSecret(pod *k8v1.PodSpec, secret string) {
	pullSecret := k8v1.LocalObjectReference{Name: secret}
	pod.ImagePullSecrets = append(pod.ImagePullSecrets, pullSecret)
}

func CreateService(name string, selector string, ports []k8v1.ServicePort) *ServiceCommand {
	s := &k8v1.Service{}
	s.Kind = "Service"
	s.APIVersion = "v1"
	s.Name = name
	if selector != "" {
		s.Spec.Selector = make(map[string]string)
		s.Spec.Selector["app"] = selector
	}
	s.Spec.Ports = ports
	serv := &ServiceCommand{Service: s}
	return serv
}
