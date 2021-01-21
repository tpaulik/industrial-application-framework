// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

/*the purpose of this module to filter every event which doesn't have any meaning for the applications.
eg: status updates, restart cases, etc*/
package consul

import (
	"reflect"

	app "github.com/nokia/industrial-application-framework/consul-operator/pkg/apis/app/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/util/finalizer"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type CustomPredicate struct{}

func (CustomPredicate) Create(event event.CreateEvent) bool {
	logger := log.WithName("predicate").WithName("create_event")
	instance, ok := event.Object.(*app.Consul)

	logger.V(1).Info("Event received", "object", instance)

	if ok {
		logger.Info("Event can be reconciled")
		return true
	}

	logger.V(1).Info("Event skipped")
	return false
}

func (CustomPredicate) Delete(event event.DeleteEvent) bool {
	logger := log.WithName("predicate").WithName("delete_event")
	logger.V(1).Info("Event received")
	instance, ok := event.Object.(*app.Consul)

	if ok && finalizer.HasFinalizers(instance) {
		logger.Info("Event can be reconciled")
		return true
	}
	logger.V(1).Info("Event skipped")
	return false
}

func (CustomPredicate) Update(event event.UpdateEvent) bool {
	logger := log.WithName("predicate").WithName("update_event")
	logger.V(1).Info("Event received")

	oldInstance, okOld := event.ObjectOld.(*app.Consul)
	newInstance, okNew := event.ObjectNew.(*app.Consul)

	if okOld && okNew {
		logger.V(1).Info("New object content", "object", newInstance)
		logger.V(1).Info("Old object content", "object", oldInstance)
		if isChangeInSpec(oldInstance, newInstance) {
			logger.Info("Event can be reconciled")
			return true
		} else if isFinalizerAddition(oldInstance, newInstance) {
			logger.Info("Finalizer added allow the reconciliation")
			return true
		} else if isDeleteEvent(oldInstance, newInstance) {
			logger.Info("DeleteTimestamp changed")
			return true
		}
	}
	logger.V(1).Info("Update event skipped")
	return false
}

func isChangeInSpec(oldInstance *app.Consul, newInstance *app.Consul) bool {
	return !reflect.DeepEqual(oldInstance.Spec, newInstance.Spec)
}

func isFinalizerAddition(oldInstance *app.Consul, newInstance *app.Consul) bool {
	return !finalizer.HasFinalizers(oldInstance) && finalizer.HasFinalizers(newInstance)
}

func isDeleteEvent(oldInstance *app.Consul, newInstance *app.Consul) bool {
	return oldInstance.DeletionTimestamp == nil && newInstance.DeletionTimestamp != nil
}

func (CustomPredicate) Generic(event.GenericEvent) bool {
	logger := log.WithName("predicate").WithName("generic event")
	logger.V(1).Info("Generic event received, skip it.")
	return false
}
