// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package v1alpha1

import (
	"errors"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/k8sdynamic"
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulSpec defines the desired state of Consul
type ConsulSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ReplicaCount         int                                `json:"replicaCount"`
	Ports                Ports                              `json:"ports"`
	PrivateNetworkAccess *common_types.PrivateNetworkAccess `json:"privateNetworkAccess,omitempty"`
}

func (in *ConsulSpec) GetPrivateNetworkAccess() *common_types.PrivateNetworkAccess {
	return in.PrivateNetworkAccess
}

type AppReportedData struct {
	//Ip addresses of the services that received IP address from the private network
	PrivateNetworkIpAddress map[string]string `json:"privateNetworkIpAddresses,omitempty"`
}

func (in *AppReportedData) SetPrivateNetworkIpAddress(privateNetworkIpAddress map[string]string) {
	in.PrivateNetworkIpAddress = privateNetworkIpAddress
}

// ConsulStatus defines the observed state of Consul
type ConsulStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	PrevSpec         *ConsulSpec                     `json:"prevSpec,omitempty"`
	AppStatus        common_types.AppStatus          `json:"appStatus,omitempty"`
	AppReportedData  AppReportedData                 `json:"appReportedData,omitempty"`
	AppliedResources []k8sdynamic.ResourceDescriptor `json:"appliedResources,omitempty"`
}

func (in *ConsulStatus) GetAppStatus() common_types.AppStatus {
	return in.AppStatus
}

func (in *ConsulStatus) SetAppStatus(status common_types.AppStatus) {
	in.AppStatus = status
}

func (in *ConsulStatus) GetPrevSpec() common_types.OperatorSpec {
	return in.PrevSpec
}

func (in *ConsulStatus) GetPrevSpecDeepCopy() common_types.OperatorSpec {
	return in.PrevSpec.DeepCopy()
}

func (in *ConsulStatus) SetPrevSpec(spec common_types.OperatorSpec) error {
	switch spec.(type) {
	case *ConsulSpec:
		consulSpec := spec.(*ConsulSpec)
		in.PrevSpec = consulSpec
		return nil
	default:
		return errors.New("SetPrevSpec type is not of type *Consulspec")
	}
}

func (in *ConsulStatus) GetAppliedResources() []k8sdynamic.ResourceDescriptor {
	return in.AppliedResources

}
func (in *ConsulStatus) SetAppliedResources(resources []k8sdynamic.ResourceDescriptor) {
	in.AppliedResources = resources
}

type Ports struct {
	UiPort    int `json:"uiPort,omitempty"`
	AltPort   int `json:"altPort,omitempty"`
	UdpPort   int `json:"udpPort,omitempty"`
	HttpPort  int `json:"httpPort,omitempty"`
	HttpsPort int `json:"httpsPort,omitempty"`
	Serflan   int `json:"serflan,omitempty"`
	Serfwan   int `json:"serfwan,omitempty"`
	ConsulDns int `json:"consulDns,omitempty"`
	Server    int `json:"server,omitempty"`
}

// +kubebuilder:object:root=true

// Consul is the Schema for the consuls API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=consuls,scope=Namespaced
// +k8s:openapi-gen=true
type Consul struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulSpec   `json:"spec,omitempty"`
	Status ConsulStatus `json:"status,omitempty"`
}

func (in *Consul) GetTypeMeta() metav1.TypeMeta {
	return in.TypeMeta
}

func (in *Consul) GetObjectMeta() metav1.ObjectMeta {
	return in.ObjectMeta
}
func (in *Consul) GetSpec() common_types.OperatorSpec {
	return &in.Spec
}
func (in *Consul) GetStatus() common_types.OperatorStatus {
	return &in.Status
}

func CreateAppInstance() common_types.OperatorCr {
	return &Consul{}
}

// +kubebuilder:object:root=true

// ConsulList contains a list of Consul
type ConsulList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Consul `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Consul{}, &ConsulList{})
}
