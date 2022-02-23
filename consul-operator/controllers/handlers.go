// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package controllers

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"reflect"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	usingPnaLabelKey = "ndac.appfw.private-network-access"
	appPnaName       = "private-network-for-consul"
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
	if appParametersChanged(instance) {
		log.V(1).Info("Network settings updated, reloading app")

		err := r.undeployAppComponentsAffectedByUpdate(namespace)
		if err != nil {
			logger.Error(err, "Failed removal of app components using pna")
			return reconcile.Result{}, nil
		}
		logger.V(1).Info("Consul app undeployed")

		//Removing the resources that are affected by the update
		pna := k8sdynamic.ResourceDescriptor{
			Name:      appPnaName,
			Namespace: namespace,
			Gvr: k8sdynamic.GroupVersionResource{
				Group:    "ops.dac.nokia.com",
				Version:  "v1alpha1",
				Resource: "privatenetworkaccesses",
			}}

		err = deleteChangedResources(pna)
		if err != nil {
			logger.Error(err, "failed to delete private network access")
			return reconcile.Result{}, err
		}

		//PrivateNetworkAccess release takes some time, we need to wait for its removal before recreating it
		waitUntilResourcesAreReleased(pna)

		//Recreate resources and redeploy application
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
		appliedPlatformResourceDescriptors, err := platformres.ApplyPnaResourceRequests(namespace)
		if err != nil {
			logger.Error(err, "failed to apply updated pna request")
			return reconcile.Result{}, nil
		}
		//Blocks until all of the platform requests granted
		err = platformres.WaitUntilResourcesGranted(appliedPlatformResourceDescriptors, time.Second*500)
		if err != nil {
			logger.Error(err, "failed to get pna resources")
			return reconcile.Result{}, nil
		}
	}

	// Upgrade application for the new networking settings to take effect
	h := helm.NewHelm(namespace)
	if err := h.Deploy(); err != nil {
		logger.Error(err, "failed to update the helm chart")
		return reconcile.Result{}, err
	}
	instance.Status.PrevSpec = &instance.Spec
	if err := r.updateStatus(instance); nil != err {
		log.Error(err, "status previous spec update failed")
	}

	return reconcile.Result{}, nil
}

func deleteChangedResources(pna k8sdynamic.ResourceDescriptor) error {
	k8sClient := k8sdynamic.New(kubelib.GetKubeAPI())
	if err := k8sClient.DeleteResources([]k8sdynamic.ResourceDescriptor{pna}); err != nil {
		return err
	}
	return nil
}

func waitUntilResourcesAreReleased(pna k8sdynamic.ResourceDescriptor) {
	logger := log.WithName("handlers").WithName("waitUntilResourcesAreReleased")
	for {
		_, err := k8sdynamic.GetDynamicK8sClient().Resource(pna.Gvr.GetGvr()).Namespace(pna.Namespace).Get(context.TODO(), pna.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				logger.V(1).Info("PNA successfully removed")
				break
			} else {
				logger.V(1).Error(err, "error getting oldpna")
			}
		}
		logger.V(1).Info("Waiting for PNA deletion")
		time.Sleep(time.Millisecond * 100)
	}
}

func (r *ConsulReconciler) undeployAppComponentsAffectedByUpdate(namespace string) error {
	//Remove statefulsets having pna label
	consulApp := &appsv1.StatefulSet{}
	opts := []client.DeleteAllOfOption{
		client.InNamespace(namespace),
		client.MatchingLabels{usingPnaLabelKey: appPnaName},
		client.GracePeriodSeconds(0),
		client.PropagationPolicy(metav1.DeletePropagationBackground),
	}
	err := r.DeleteAllOf(context.TODO(), consulApp, opts...)
	return err
}

func appParametersChanged(instance *app.Consul) bool {
	// Comparing existing application parameters with new values in case of parameters whose value change is supported
	return !reflect.DeepEqual(instance.Status.PrevSpec.PrivateNetworkAccess, instance.Spec.PrivateNetworkAccess)
}

func (r *ConsulReconciler) updateStatus(instance *app.Consul) error {
	prevSpec := instance.Status.PrevSpec.DeepCopy()
	appliedResources := make([]k8sdynamic.ResourceDescriptor, len(instance.Status.AppliedResources))
	copy(appliedResources, instance.Status.AppliedResources)
	key := client.ObjectKey{
		Namespace: instance.GetNamespace(),
		Name:      instance.GetName(),
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := r.Get(context.TODO(), key, instance)
		if err != nil {
			return err
		}

		instance.Status.PrevSpec = prevSpec
		instance.Status.AppliedResources = appliedResources

		err = r.Status().Update(context.TODO(), instance)
		return err
	})

	if err != nil {
		return errors.Wrap(err, "failed Consul status update")
	}

	return nil
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
					appPnaName,
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

	pnaObj, err := getPna(namespace, pnaName, k8sClient)
	if err != nil {
		logger.Error(err, "Failed to get the PrivateNetworkAccess CR")
		return nil
	}
	assignedNetwork, found, _ := unstructured.NestedStringMap(pnaObj.Object, "status", "assignedNetwork")
	if found && assignedNetwork != nil {
		logger.V(1).Info("Assigned network found in status, using dummy interface")
		return getAddressOfDummyInterface(namespace, deploymentList, k8sClient)
	} else {
		pnaNetworkName, found, _ := unstructured.NestedString(pnaObj.Object, "status", "appNetworkName")
		if !found {
			logger.Error(err, "Failed to get the interface name in the PrivateNetworkAccess CR")
			return nil
		}
		return getAddressesOfPnaDefinedInterfaces(namespace, deploymentList, k8sClient, pnaNetworkName)
	}
}

func getPna(namespace string, pnaName string, k8sClient dynamic.Interface) (*unstructured.Unstructured, error) {
	pnaGvr := schema.GroupVersionResource{Version: "v1alpha1", Group: "ops.dac.nokia.com", Resource: "privatenetworkaccesses"}
	pnaObj, err := k8sClient.Resource(pnaGvr).Namespace(namespace).Get(context.TODO(), pnaName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pnaObj, err
}

func getAddressesOfPnaDefinedInterfaces(namespace string, deploymentList []deploymentId, k8sClient dynamic.Interface, pnaNetworkName string) map[string]string {
	logger := log.WithName("getAddressesOfPnaDefinedInterfaces")
	logger.V(1).Info("Read address of interface defined in PNA")
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
				retIpAddresses[string(deployment.deploymentType)+"/"+deployment.name] = netAnn.IPRequest[0]
			}
		}
	}

	return retIpAddresses
}

func getAddressOfDummyInterface(namespace string, deploymentList []deploymentId, k8sClient dynamic.Interface) map[string]string {
	logger := log.WithName("getAddressOfDummyInterface")
	retIpAddresses := make(map[string]string)
	for _, deployment := range deploymentList {
		deploymentGvr := schema.GroupVersionResource{Version: "v1", Group: "apps", Resource: string(deployment.deploymentType)}
		deploymentObj, err := k8sClient.Resource(deploymentGvr).Namespace(namespace).Get(context.TODO(), deployment.name, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "Failed to get the following deployment", "type", deployment.deploymentType, "name", deployment.name)
			break
		}
		initContainers, found, err := unstructured.NestedSlice(deploymentObj.Object, "spec", "template", "spec", "initContainers")
		if !found || err != nil {
			logger.Error(err, "Failed to read initContainers", "type", deployment.deploymentType, "name", deployment.name)
			break
		}
		for _, initContainer := range initContainers {
			logger.Info("name:" + initContainer.(map[string]interface{})["name"].(string))
			if initContainer.(map[string]interface{})["name"] == "appfw-private-network-routing" {
				if args := initContainer.(map[string]interface{})["args"]; args != "" {
					rg, err := regexp.Compile(`ip\s*link\s*add\s*name\s*.*?\s*type\s*dummy\s*&&\s*ip\s*addr\s*add\s*(?P<customerIP>.*?)/32`)
					if err != nil {
						logger.Error(err, "failed to compile the regular expression")
						break
					}
					result := rg.FindStringSubmatch(args.([]interface{})[0].(string))
					if result != nil {
						logger.Info("Found IP to use from dummy interface " + result[1])
						retIpAddresses[string(deployment.deploymentType)+"/"+deployment.name] = result[1]
					}
				} else {
					logger.Error(nil, "Failed to read init container args", "type", deployment.deploymentType, "name", deployment.name)
				}
			}
		}
	}
	return retIpAddresses
}
