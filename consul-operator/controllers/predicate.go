// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

/*the purpose of this module to filter every event which doesn't have any meaning for the applications.
eg: status updates, restart cases, etc*/
package controllers

import (
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/util/finalizer"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
)

type CustomPredicate struct{}

func (CustomPredicate) Create(event event.CreateEvent) bool {
	logger := log.WithName("predicate").WithName("create_event")
	instance, ok := event.Object.(common_types.OperatorCr)

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
	instance, ok := event.Object.(common_types.OperatorCr)

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

	oldInstance, okOld := event.ObjectOld.(common_types.OperatorCr)
	newInstance, okNew := event.ObjectNew.(common_types.OperatorCr)

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

func isChangeInSpec(oldInstance common_types.OperatorCr, newInstance common_types.OperatorCr) bool {
	return !reflect.DeepEqual(oldInstance.GetSpec(), newInstance.GetSpec())
}

func isFinalizerAddition(oldInstance common_types.OperatorCr, newInstance common_types.OperatorCr) bool {
	return !finalizer.HasFinalizers(oldInstance) && finalizer.HasFinalizers(newInstance)
}

func isDeleteEvent(oldInstance common_types.OperatorCr, newInstance common_types.OperatorCr) bool {
	return oldInstance.GetDeletionTimestamp() == nil && newInstance.GetDeletionTimestamp() != nil
}

func (CustomPredicate) Generic(event.GenericEvent) bool {
	logger := log.WithName("predicate").WithName("generic event")
	logger.V(1).Info("Generic event received, skip it.")
	return false
}
