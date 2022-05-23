// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package controllers

import (
	"context"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/handlers"
	app "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_consul")

// OperatorReconciler reconciles a Application object
type ConsulReconciler struct {
	Common handlers.OperatorReconciler
}

//<kubebuilder-annotations>
//+kubebuilder:rbac:groups=app.dac.nokia.com,namespace=app-ns,resources=consuls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.dac.nokia.com,namespace=app-ns,resources=consuls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.dac.nokia.com,namespace=app-ns,resources=consuls/finalizers,verbs=update
//+kubebuilder:rbac:groups=ops.dac.nokia.com,namespace=app-ns,resources=*,verbs=create;delete;get;list;patch;update;watch
//+kubebuilder:rbac:groups="extensions;networking.k8s.io",namespace=app-ns,resources=ingresses,verbs=*
//+kubebuilder:rbac:groups="",namespace=app-ns,resources=pods;services;endpoints;events;configmaps;secrets,verbs=create;delete;get;list;watch;patch;update
//+kubebuilder:rbac:groups="apps",namespace=app-ns,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
//+kubebuilder:rbac:groups="apps",resourceNames=consul-operator,namespace=app-ns,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="monitoring.coreos.com",namespace=app-ns,resources=servicemonitors,verbs=get;create
//+kubebuilder:rbac:groups="coordination.k8s.io",namespace=app-ns,resources=leases,verbs=get;list;watch;create;update;patch;delete
//</kubebuilder-annotations>

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Consul object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-etcd
//kube-apiserverruntime@v0.9.2/pkg/reconcile
func (r *ConsulReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	return r.Common.Reconcile(ctx, request)
}

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr ctrl.Manager, reconciler reconcile.Reconciler) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(app.CreateAppInstance()).
		Owns(&corev1.Pod{}).
		WithEventFilter(&CustomPredicate{}).
		Complete(reconciler)
}
