// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package licenceexpired

import (
	"context"
	"github.com/nokia/industrial-application-framework/alarmlogger"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/licenceexpired"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/monitoring"
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("licence_expired_handler")

func CreateLicenseExpiredHandler(runtimeClient client.Client, appInstance common_types.OperatorCr, clientSet *kubernetes.Clientset, monitor *monitoring.Monitor) licenceexpired.LicenceExpiredResourceFuncs {
	return &SampleFuncs{
		RuntimeClient: runtimeClient,
		AppInstance:   appInstance,
		ClientSet:     clientSet,
		Monitor:       monitor}
}

// BEGIN sample callback functions
type SampleFuncs struct {
	RuntimeClient client.Client
	AppInstance   common_types.OperatorCr
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
	cb.AppInstance.GetStatus().SetAppStatus(common_types.AppStatusFrozen)
	if err := cb.RuntimeClient.Status().Update(context.TODO(), cb.AppInstance); nil != err {
		log.Error(err, "status appStatus update failed", "appStatus", cb.AppInstance.GetStatus().GetAppStatus())
	}

	ns := cb.AppInstance.GetObjectMeta().Namespace
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

	ns := cb.AppInstance.GetObjectMeta().Namespace
	cb.AppInstance.GetStatus().SetAppStatus(cb.Monitor.GetApplicationStatus())
	if err := cb.RuntimeClient.Status().Update(context.TODO(), cb.AppInstance); nil != err {
		log.Error(err, "status appStatus update failed", "appStatus", cb.AppInstance.GetStatus().GetAppStatus())
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
