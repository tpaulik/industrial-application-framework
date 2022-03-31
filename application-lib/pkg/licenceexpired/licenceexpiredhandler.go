// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package licenceexpired

import (
	"github.com/nokia/industrial-application-framework/application-lib/pkg/k8sdynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
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
