// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package v1alpha1

import (
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/k8sdynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AppStatus string

const (
	AppStatusNotSet     = "UNSET"
	AppStatusNotRunning = "NOT_RUNNING"
	AppStatusRunning    = "RUNNING"
	AppStatusFrozen     = "FROZEN"
)

// ConsulSpec defines the desired state of Consul
// +k8s:openapi-gen=true
type ConsulSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: RunCrTemplater "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ReplicaCount int   `json:"replicaCount"`
	Ports        Ports `json:"ports"`
}

type AppReporteData struct {
	//The structure of this type is up the application. AppFw will convert the whole representation to JSON.
	MetricsClusterIp string `json:"metricsClusterIp,omitempty"`
}

// ConsulStatus defines the observed state of Consul
// +k8s:openapi-gen=true
type ConsulStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: RunCrTemplater "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	PrevSpec         *ConsulSpec                     `json:"prevSpec,omitempty"`
	AppStatus        AppStatus                       `json:"appStatus,omitempty"`
	AppReportedData  AppReporteData                  `json:"appReportedData,omitempty"`
	AppliedResources []k8sdynamic.ResourceDescriptor `json:"appliedResources,omitempty"`
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Consul is the Schema for the consuls API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Consul struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulSpec   `json:"spec,omitempty"`
	Status ConsulStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsulList contains a list of Consul
type ConsulList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Consul `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Consul{}, &ConsulList{})
}
