// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package common_types

import (
	"github.com/nokia/industrial-application-framework/application-lib/pkg/k8sdynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type OperatorCr interface {
	GetTypeMeta() metav1.TypeMeta
	GetObjectMeta() metav1.ObjectMeta
	GetSpec() OperatorSpec
	GetStatus() OperatorStatus
	metav1.Object
	runtime.Object
}

type OperatorSpec interface {
	GetPrivateNetworkAccess() *PrivateNetworkAccess
}

type OperatorStatus interface {
	SetAppStatus(status AppStatus)
	GetAppStatus() AppStatus
	GetAppliedResources() []k8sdynamic.ResourceDescriptor
	SetAppliedResources(resources []k8sdynamic.ResourceDescriptor)
	GetPrevSpec() OperatorSpec
	GetPrevSpecDeepCopy() OperatorSpec
	SetPrevSpec(spec OperatorSpec) error
}
