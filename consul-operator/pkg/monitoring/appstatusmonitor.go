// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package monitoring

import (
	"context"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/handlers"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/monitoring"
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	"github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("monitoring_consul")

func CreateAppStatusMonitor(instance common_types.OperatorCr, namespace string, reconciler *handlers.OperatorReconciler) *monitoring.Monitor {
	logger := log.WithName("controller").WithName("createAppStatusMonitor").WithValues("namespace", namespace, "name", instance.GetObjectMeta().Name)
	logger.Info("Called")

	appStatusMonitor := monitoring.NewMonitor(reconciler.Client, instance, namespace,
		func() {
			logger.Info("Set AppReportedData")
			//runningCallback - example, some dynamic data should be reported here which has value only after the deployment

			consulInstance := instance.(*v1alpha1.Consul)

			//not necessarily needed
			if instance.GetSpec().GetPrivateNetworkAccess() != nil {
				consulInstance.Status.AppReportedData.PrivateNetworkIpAddress = handlers.GetPrivateNetworkIpAddresses(
					namespace,
					reconciler.Configuration.AppPnaName,
					[]handlers.DeploymentId{
						{DeploymentType: handlers.DeploymentTypeStatefulset,
							Name: reconciler.Configuration.DeploymentName},
					},
				)
			}

			if err := reconciler.Client.Status().Update(context.TODO(), instance); nil != err {
				logger.Error(err, "status app reported data update failed")
			}
		},
		func() {
			//notRunningCallback
		},
	)

	return appStatusMonitor
}
