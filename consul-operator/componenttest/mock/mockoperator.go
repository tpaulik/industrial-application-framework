package mock

import (
	"errors"
	"k8s.io/client-go/tools/cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var MockOperators = make(map[string]*MockOperator, 0)
var FwLog = logf.Log.WithName("componenttest")

type MockOperator struct {
	name         string
	stopper      chan struct{}
	Informer     cache.SharedIndexInformer
	eventHandler cache.ResourceEventHandlerFuncs
}

func StartMockOperators() {
	for _, mockService := range MockOperators {
		go mockService.Start()
	}
}

func StopMockOperators() {
	for name, mockService := range MockOperators {
		mockService.Stop()
		delete(MockOperators, name)
	}
}

func RunMockOperator(name string) error {
	if _, exists := MockOperators[name]; !exists {
		return errors.New("Unable to start MockOperator. " + name + " not found")
	}
	go MockOperators[name].Start()
	return nil
}

func StopMockOperator(name string) error {
	if _, exists := MockOperators[name]; !exists {
		return errors.New("Unable to stop MockOperator. " + name + " not found")
	}
	MockOperators[name].Stop()
	delete(MockOperators, name)
	return nil
}

func NewMockOperator(serviceName string) (mockService *MockOperator) {
	mockService = new(MockOperator)
	mockService.name = serviceName
	mockService.stopper = make(chan struct{})
	mockService.eventHandler = cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { FwLog.Info(mockService.name + ": Add not implemented") },
		DeleteFunc: func(obj interface{}) { FwLog.Info(mockService.name + ": Delete not implemented") },
		UpdateFunc: func(oldObj interface{}, newObj interface{}) { FwLog.Info(mockService.name + ": Update not implemented") },
	}
	MockOperators[serviceName] = mockService
	return
}

func (m *MockOperator) Start() {
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

func (m *MockOperator) HandleAdd(handler func(obj interface{})) {
	m.eventHandler.AddFunc = handler
}

func (m *MockOperator) HandleUpdate(handler func(oldObj interface{}, newObj interface{})) {
	m.eventHandler.UpdateFunc = handler
}

func (m *MockOperator) HandleDelete(handler func(obj interface{})) {
	m.eventHandler.DeleteFunc = handler
}

func (m *MockOperator) Stop() {
	close(m.stopper)
}
