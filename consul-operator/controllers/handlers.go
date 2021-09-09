// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package controllers

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"time"

	netattv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	app "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/libs/kubelib"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/helm"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/k8sdynamic"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/licenceexpired"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/monitoring"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/platformres"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/template"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/util/finalizer"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	deploymentTypeDeployment  = "deployments"
	deploymentTypeStatefulset = "statefulsets"
	deploymentTypeDaemonset   = "deamonsets"
)

type deploymentType string

type deploymentId struct {
	deploymentType deploymentType
	name           string
}

var appStatusMonitor *monitoring.Monitor

func (r *ConsulReconciler) handleCrChange(instance *app.Consul, namespace string) (reconcile.Result, error) {
	logger := log.WithName("handlers").WithName("handleCrChange").WithValues("namespace", namespace, "name", instance.ObjectMeta.Name)
	logger.Info("Event arrived handle it")
	if instance.ObjectMeta.DeletionTimestamp != nil {
		//Object should be deleted
		return r.handleDelete(instance, namespace)
	}

	if !finalizer.HasFinalizers(instance) {
		logger.Info("Add finalizer")
		err := finalizer.AddFinalizer(instance, finalizer.FinalizerId)
		if err != nil {
			logger.Error(err, "Failed to set finalizer")
			return reconcile.Result{}, err
		}
		err = r.Client.Update(context.TODO(), instance)
		return reconcile.Result{}, nil
	}

	if isSpecUpdated(instance) {
		return r.handleUpdate(instance, namespace)
	} else {
		return r.handleCreate(instance, namespace)
	}
}

func isSpecUpdated(instance *app.Consul) bool {
	return instance.Status.PrevSpec != nil && !reflect.DeepEqual(instance.Spec, *instance.Status.PrevSpec)
}

func (r *ConsulReconciler) handleDelete(instance *app.Consul, namespace string) (reconcile.Result, error) {
	logger := log.WithName("handlers").WithName("handleDelete").WithValues("namespace", namespace, "name", instance.ObjectMeta.Name)
	logger.Info("Called")

	if nil != appStatusMonitor {
		appStatusMonitor.Pause()
	}

	//Go through the app spec CR and delete all of the resources present in the AppliedResources list
	k8sClient := k8sdynamic.New(kubelib.GetKubeAPI())
	if err := k8sClient.DeleteResources(instance.Status.AppliedResources); err != nil {
		logger.Error(err, "failed to delete the resources")
	}

	//Optional - if helm was used for the deployment helm has to be used also for the undeployment
	h := helm.NewHelm(namespace)
	if err := h.Undeploy(); err != nil {
		logger.Error(err, "failed to uninstall the helm chart")
	}

	finalizer.RemoveFinalizer(instance, finalizer.FinalizerId)
	r.Client.Update(context.TODO(), instance)

	return reconcile.Result{}, nil
}

func (r *ConsulReconciler) handleUpdate(instance *app.Consul, namespace string) (reconcile.Result, error) {
	logger := log.WithName("handlers").WithName("handleUpdate").WithValues("namespace", namespace, "name", instance.ObjectMeta.Name)
	logger.Info("Called")

	instance.Status.PrevSpec = &instance.Spec
	if err := r.Client.Status().Update(context.TODO(), instance); nil != err {
		log.Error(err, "status previous spec update failed")
	}

	return reconcile.Result{}, nil
}

func (r *ConsulReconciler) handleCreate(instance *app.Consul, namespace string) (reconcile.Result, error) {
	logger := log.WithName("handlers").WithName("handleCreate").WithValues("namespace", namespace, "name", instance.ObjectMeta.Name)
	logger.Info("Called")

	//Execute CR based templating to resolve the variables in the resource-req dir
	resReqTemplater, err := template.NewTemplater(instance.Spec, namespace, "resource-reqs")
	if err != nil {
		logger.Error(err, "Failed to initialize the res-req appDeploymentTemplater")
		return reconcile.Result{}, nil
	}
	_, err = resReqTemplater.RunCrTemplater("---\n")
	if err != nil {
		logger.Error(err, "Failed to execute the res-req CR appDeploymentTemplater")
		return reconcile.Result{}, nil
	}

	//Request NDAC platform resources
	appliedPlatformResourceDescriptors, err := platformres.ApplyPlatformResourceRequests(namespace)
	if err != nil {
		logger.Error(err, "failed to apply the platform resource requests")
		return reconcile.Result{}, nil
	}
	//Blocks until all of the platform requests granted
	err = platformres.WaitUntilResourcesGranted(appliedPlatformResourceDescriptors, time.Second*500)
	if err != nil {
		logger.Error(err, "failed to get all of the requested platform resources")
		return reconcile.Result{}, nil
	}

	//Execute templating for the app-deplyoment directory using the values from the CR
	appDeploymentTemplater, err := template.NewTemplater(instance.Spec, namespace, "app-deployment")
	if err != nil {
		logger.Error(err, "Failed to initialize the appDeploymentTemplater")
		return reconcile.Result{}, nil
	}

	//Gives back the output of the templated yamls. It can be applied directly in kubernetes
	_, err = appDeploymentTemplater.RunCrTemplater("---\n")
	if err != nil {
		logger.Error(err, "Failed to execute the CR appDeploymentTemplater")
		return reconcile.Result{}, nil
	}

	//Optional - Helm based deployment
	err = helm.NewHelm(namespace).Deploy()
	if err != nil {
		logger.Error(err, "Failed to deploy the helm chart")
		return reconcile.Result{}, nil
	}

	//This section is only needed if helm is not used for the deployment
	//k8sClient := k8sdynamic.New(kubelib.GetKubeAPI())
	//appliedApplicationResourceDescriptors, err := k8sClient.ApplyConcatenatedResources(out, namespace)
	//if err != nil {
	//	logger.Error(err, "failed to apply the templated resources")
	//	return reconcile.Result{}, nil
	//}

	instance.Status.AppliedResources = appliedPlatformResourceDescriptors
	//This section is only needed if helm is not used for the deployment
	//instance.Status.AppliedResources = append(instance.Status.AppliedResources, appliedApplicationResourceDescriptors...)
	instance.Status.PrevSpec = &instance.Spec
	if err := r.Client.Status().Update(context.TODO(), instance); nil != err {
		logger.Error(err, "status applied resources and previous spec update failed")
	}

	//Controls the appStatus and appReportedData in the app spec CR, running continuously in the background
	appStatusMonitor = monitoring.NewMonitor(r.Client, instance, namespace,
		func() {
			logger.Info("Set AppReportedData")
			//runningCallback - example, some dynamic data should be reported here which has value only after the deployment
			svc, err := kubelib.GetKubeAPI().CoreV1().Services(namespace).Get(context.TODO(), "example-consul-service", metav1.GetOptions{})
			if err != nil {
				logger.Error(err, "Failed to read the svc of the metrics endpoint")
				return
			}
			instance.Status.AppReportedData.MetricsClusterIp = svc.Spec.ClusterIP
			if instance.Spec.PrivateNetworkAccess != nil {
				instance.Status.AppReportedData.PrivateNetworkIpAddress = getPrivateNetworkIpAddresses(
					namespace,
					"private-network-for-consul",
					[]deploymentId{
						{deploymentTypeStatefulset, "example-consul"},
					},
				)
			}

			if err := r.Client.Status().Update(context.TODO(), instance); nil != err {
				logger.Error(err, "status app reported data update failed")
			}
		},
		func() {
			//notRunningCallback
		},
	)
	appStatusMonitor.Run()

	//Handles the application license expiration, reactivation
	licCallbacks := &licenceexpired.SampleFuncs{
		RuntimeClient: r.Client,
		AppInstance:   instance,
		ClientSet:     kubelib.GetKubeAPI(),
		Monitor:       appStatusMonitor,
	}

	licenceexpired.New(namespace, licCallbacks).Watch()

	return reconcile.Result{}, nil
}

func getPrivateNetworkIpAddresses(namespace, pnaName string, deploymentList []deploymentId) map[string]string {
	logger := log.WithName("getPrivateNetworkIpAddresses")
	k8sClient := k8sdynamic.GetDynamicK8sClient()

	pnaGvr := schema.GroupVersionResource{Version: "v1alpha1", Group: "ops.dac.nokia.com", Resource: "privatenetworkaccesses"}
	pnaObj, err := k8sClient.Resource(pnaGvr).Namespace(namespace).Get(context.TODO(), pnaName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Failed to get the PrivateNetworkAccess CR")
		return nil
	}
	pnaNetworkName, found, _ := unstructured.NestedString(pnaObj.Object, "status", "appNetworkName")
	if !found {
		logger.Error(err, "Failed to get the interface name in the PrivateNetworkAccess CR")
		return nil
	}

	retIpAddresses := make(map[string]string)
	for _, deployment := range deploymentList {
		deploymentGvr := schema.GroupVersionResource{Version: "v1", Group: "apps", Resource: string(deployment.deploymentType)}
		deploymentObj, err := k8sClient.Resource(deploymentGvr).Namespace(namespace).Get(context.TODO(), deployment.name, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "Failed to get the following deployment", "type", deployment.deploymentType, "name", deployment.name)
			break
		}
		value, found, _ := unstructured.NestedString(deploymentObj.Object, "spec", "template", "metadata", "annotations", "k8s.v1.cni.cncf.io/networks")
		if !found {
			logger.Error(nil, "Failed to get the assigned IP from the deployment", "type", deployment.deploymentType, "name", deployment.name)
			break
		}
		var parsedNetAnn []netattv1.NetworkSelectionElement
		err = json.Unmarshal([]byte(value), &parsedNetAnn)
		if err != nil {
			logger.Error(err, "Failed to parse the current network annotation in the deployment", "type", deployment.deploymentType, "name", deployment.name)
			break
		}
		for _, netAnn := range parsedNetAnn {
			if netAnn.Name == pnaNetworkName {
				retIpAddresses[string(deployment.deploymentType)+"/"+deployment.name] = netAnn.IPRequest[0] //TODO
			}
		}
	}

	return retIpAddresses
}
