package controller

import (
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/controller/consul"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a controllerextension.
	AddToManagerFuncs = append(AddToManagerFuncs, consul.Add)
}
