// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package kubelib

import (
	log "github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	k8v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type Command interface{}

type KubeCommand interface {
	Add(clientset kubernetes.Interface) error
	Undo(clientset kubernetes.Interface) error
	Update(clientset kubernetes.Interface) error
}

type ConfigMapCommand struct {
	Configmap *k8v1.ConfigMap
}

type DeploymentCommand struct {
	Deployment *apps.Deployment
}

type ServiceCommand struct {
	Service *k8v1.Service
}

type JobCommand struct {
	Job *batch.Job
}

type KubernetesCommand struct {
	Obj *runtime.Object
}

func (d *DeploymentCommand) Add(clientset kubernetes.Interface) error {
	_, err := clientset.AppsV1beta1().Deployments("default").Create(d.Deployment)
	return err
}
func (d *DeploymentCommand) Undo(clientset kubernetes.Interface) error {
	err := clientset.AppsV1beta1().Deployments("default").Delete(d.Deployment.Name, nil)
	return err
}
func (d *DeploymentCommand) Update(clientset kubernetes.Interface) error {
	_, err := clientset.AppsV1beta1().Deployments("default").Update(d.Deployment)
	return err
}

func (j *JobCommand) Add(clientset kubernetes.Interface) error {
	_, err := clientset.BatchV1().Jobs("default").Create(j.Job)
	return err
}

func (j *JobCommand) Undo(clientset kubernetes.Interface) error {
	err := clientset.BatchV1().Jobs("default").Delete(j.Job.Name, nil)
	return err
}
func (j *JobCommand) Update(clientset kubernetes.Interface) error {
	_, err := clientset.BatchV1().Jobs("default").Update(j.Job)
	return err
}
func (s *ServiceCommand) Add(clientset kubernetes.Interface) error {
	_, err := clientset.CoreV1().Services("default").Create(s.Service)
	return err

}
func (s *ServiceCommand) Update(clientset kubernetes.Interface) error {
	return nil
}
func (s *ServiceCommand) Undo(clientset kubernetes.Interface) error {
	err := clientset.CoreV1().Services("default").Delete(s.Service.Name, nil)
	return err
}

func (k *ConfigMapCommand) Add(clientset kubernetes.Interface) error {
	_, err := clientset.CoreV1().ConfigMaps("default").Create(k.Configmap)
	return err
}

func (k *ConfigMapCommand) Update(clientset kubernetes.Interface) error {
	_, err := clientset.CoreV1().ConfigMaps("default").Update(k.Configmap)
	if err != nil {
		log.Errorf("Failed to update configmap %s ", err.Error())
	}
	return err
}

//Undo deletes the ConfigMap from k8s
func (k *ConfigMapCommand) Undo(clientset kubernetes.Interface) error {
	err := clientset.CoreV1().ConfigMaps("default").Delete(k.Configmap.Name, nil)
	if err != nil {
		log.Errorf("Failed to update configmap %s ", err.Error())
	}
	return err
}
