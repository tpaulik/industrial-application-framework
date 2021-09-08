// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package licenceexpired

import (
	"context"

	"github.com/nokia/industrial-application-framework/alarmlogger"
	app "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/k8sdynamic"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/monitoring"

	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	Group    = "ops.dac.nokia.com"
	Version  = "v1alpha1"
	Resource = "licenceexpireds"
)

var (
	log     = logf.Log.WithName("licence_expired_handler")
	handler *Handler
)

type LicenceExpiredResourceFuncs interface {
	Expired()
	Activate()
}

type Handler struct {
	namespace string
	gvr       *schema.GroupVersionResource
	callbacks LicenceExpiredResourceFuncs
	watching  bool
}

func New(namespace string, callbacks LicenceExpiredResourceFuncs) *Handler {
	if nil == handler {
		handler = &Handler{
			namespace: namespace,
			gvr: &schema.GroupVersionResource{
				Group:    Group,
				Version:  Version,
				Resource: Resource,
			},
			callbacks: callbacks,
		}
	}

	return handler
}

func (h *Handler) Watch() {
	if true == h.watching {
		return
	}
	h.watching = true

	log.Info("Watch LicenceExpired Resource in ", "namespace", h.namespace)

	stopper := make(chan struct{})

	go k8sdynamic.WatchInformer("", h.namespace, "", *h.gvr,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				h.callbacks.Expired()
			},
			DeleteFunc: func(obj interface{}) {
				h.callbacks.Activate()
			},
		},
		stopper)

}

// BEGIN sample callback functions
type SampleFuncs struct {
	RuntimeClient client.Client
	AppInstance   *app.Consul
	ClientSet     *kubernetes.Clientset
	Monitor       *monitoring.Monitor
	services      []*corev1.Service
}

func (cb *SampleFuncs) Expired() {
	log.Info("Expired")

	alarmlogger.RaiseAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
		Name:     "LicenceExpired",
		ID:       "2",
		Severity: alarmlogger.Warning,
		Text:     "Application licence is invalid",
	})

	cb.Monitor.Pause()
	cb.AppInstance.Status.AppStatus = app.AppStatusFrozen
	if err := cb.RuntimeClient.Status().Update(context.TODO(), cb.AppInstance); nil != err {
		log.Error(err, "status appStatus update failed", "appStatus", cb.AppInstance.Status.AppStatus)
	}

	ns := cb.AppInstance.GetObjectMeta().GetNamespace()
	svcList, err := cb.ClientSet.CoreV1().Services(ns).List(context.TODO(), cb.getSvcListOptions())
	if nil != err {
		log.Error(err, "Failed in listing services in ", "namespace", ns)
		return
	}

	svcs := []*corev1.Service{}
	for _, svc := range svcList.Items {
		name := svc.GetObjectMeta().GetName()
		toSave := &corev1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:   name,
				Labels: svc.GetObjectMeta().GetLabels(),
			},
			Spec: svc.Spec,
		}
		svcs = append(svcs, toSave)
		deletePolicy := v1.DeletePropagationBackground
		if err := cb.ClientSet.CoreV1().Services(ns).Delete(context.TODO(), name, v1.DeleteOptions{
			PropagationPolicy: &deletePolicy}); nil != err {
			log.Error(err, "Failed to delete ", "service", name)
			continue
		}
		log.Info("Deleted ", "service", name)
	}

	cb.services = svcs
}

func (cb *SampleFuncs) getSvcListOptions() v1.ListOptions {
	listOp := v1.ListOptions{
		LabelSelector: "deleteOnLicenceExpiration=true",
	}

	return listOp
}

func (cb *SampleFuncs) Activate() {
	log.Info("Activate")

	alarmlogger.ClearAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
		Name:     "LicenceExpired",
		ID:       "2",
		Severity: alarmlogger.Warning,
		Text:     "Application licence is valid",
	})

	ns := cb.AppInstance.GetObjectMeta().GetNamespace()
	cb.AppInstance.Status.AppStatus = cb.Monitor.GetApplicationStatus()
	if err := cb.RuntimeClient.Status().Update(context.TODO(), cb.AppInstance); nil != err {
		log.Error(err, "status appStatus update failed", "appStatus", cb.AppInstance.Status.AppStatus)
	}
	cb.Monitor.Run()

	for _, svc := range cb.services {
		result, err := cb.ClientSet.CoreV1().Services(ns).Create(context.TODO(), svc, v1.CreateOptions{})
		if nil != err {
			log.Error(err, "Failed to create ", "service", svc.GetObjectMeta().GetName())
			continue
		}
		log.Info("Created ", "service", result.GetObjectMeta().GetName())
	}
	cb.services = []*corev1.Service{}
}

// END sample callback functions
