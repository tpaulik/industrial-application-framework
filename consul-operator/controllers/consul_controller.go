/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/operator-framework/operator-lib/handler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	runtimeHandler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	app "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
)

var log = logf.Log.WithName("controller_consul")

// ConsulReconciler reconciles a Consul object
type ConsulReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.dac.nokia.com,resources=consuls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.dac.nokia.com,resources=consuls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.dac.nokia.com,resources=consuls/finalizers,verbs=update
//+kubebuilder:rbac:groups=ops.dac.nokia.com,resources=*,verbs=create;delete;get;list;patch;update;watch
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=*
//+kubebuilder:rbac:groups="",resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=create;delete;get;list;watch;patch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Consul object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ConsulReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Consul")

	// Fetch the Consul instance
	instance := &app.Consul{}
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

	return r.handleCrChange(instance, request.Namespace)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsulReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new controller
	c, err := controller.New("consul-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Consul
	err = c.Watch(&source.Kind{Type: &app.Consul{}}, &handler.InstrumentedEnqueueRequestForObject{}, &CustomPredicate{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Consul
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &runtimeHandler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &app.Consul{},
	})
	if err != nil {
		return err
	}

	return nil
}
