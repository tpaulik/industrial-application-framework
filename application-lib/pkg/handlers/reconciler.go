// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package handlers

import (
	"context"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/config"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/licenceexpired"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/monitoring"
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("common_operator_controller")

type AppCrCreator func() common_types.OperatorCr
type MonitorInstanceCreator func(common_types.OperatorCr, string, *OperatorReconciler) *monitoring.Monitor
type LicenceExpiredHandlerCreator func(client.Client, common_types.OperatorCr, *kubernetes.Clientset, *monitoring.Monitor) licenceexpired.LicenceExpiredResourceFuncs
type AppParametersChangedChecker func(common_types.OperatorCr) bool

type ReconcilerHookFunctions struct {
	CreateAppCr                 AppCrCreator
	CreateAppStatusMonitor      MonitorInstanceCreator
	CreateLicenceExpiredHandler LicenceExpiredHandlerCreator
	CheckAppParametersChanged   AppParametersChangedChecker
}

// OperatorReconciler reconciles a Application object
type OperatorReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Configuration config.OperatorConfig
	Functions     ReconcilerHookFunctions
}

func (r *OperatorReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling " + r.Configuration.ApplicationName)

	// Fetch the Application CR instance
	instance := r.Functions.CreateAppCr()
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return r.HandleCrChange(instance, request.Namespace)
}
