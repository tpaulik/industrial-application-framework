/*
Copyright 2020 Nokia
Licensed under the BSD 3-Clause License.
SPDX-License-Identifier: BSD-3-Clause
*/

package mock

import (
	"errors"
	"k8s.io/client-go/tools/cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var MockServices = make(map[string]*MockService, 0)
var FwLog = logf.Log.WithName("mock")

type MockService struct {
	name         string
	stopper      chan struct{}
	Informer     cache.SharedIndexInformer
	eventHandler cache.ResourceEventHandlerFuncs
}

func StartMockServices() {
	for _, mockService := range MockServices {
		go mockService.Start()
	}
}

func StopMockServices() {
	for name, mockService := range MockServices {
		mockService.Stop()
		delete(MockServices, name)
	}
}

func RunMockService(name string) error {
	if _, exists := MockServices[name]; !exists {
		return errors.New("Unable to start MockService. " + name + " not found")
	}
	go MockServices[name].Start()
	return nil
}

func StopMockService(name string) error {
	if _, exists := MockServices[name]; !exists {
		return errors.New("Unable to stop MockService. " + name + " not found")
	}
	MockServices[name].Stop()
	delete(MockServices, name)
	return nil
}

func NewMockService(serviceName string) (mockService *MockService) {
	mockService = new(MockService)
	mockService.name = serviceName
	mockService.stopper = make(chan struct{})
	mockService.eventHandler = cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { FwLog.Info(mockService.name + ": Add not implemented") },
		DeleteFunc: func(obj interface{}) { FwLog.Info(mockService.name + ": Delete not implemented") },
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			FwLog.Info(mockService.name + ": Update not implemented")
		},
	}
	MockServices[serviceName] = mockService
	return
}

func (m *MockService) Start() {
	FwLog.Info("Starting mock service " + m.name)
	stopCh := make(chan struct{})
	m.Informer.AddEventHandler(m.eventHandler)
	go m.Informer.Run(stopCh)
	select {
	case <-m.stopper:
		close(stopCh)
		FwLog.Info(m.name + " mock service stopped gracefully")
	}
}

func (m *MockService) HandleAdd(handler func(obj interface{})) {
	m.eventHandler.AddFunc = handler
}

func (m *MockService) HandleUpdate(handler func(oldObj interface{}, newObj interface{})) {
	m.eventHandler.UpdateFunc = handler
}

func (m *MockService) HandleDelete(handler func(obj interface{})) {
	m.eventHandler.DeleteFunc = handler
}

func (m *MockService) Stop() {
	close(m.stopper)
}
