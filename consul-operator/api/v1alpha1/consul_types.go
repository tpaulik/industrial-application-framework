// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package v1alpha1

import (
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/k8sdynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppStatus string

const (
	AppStatusNotSet     = "UNSET"
	AppStatusNotRunning = "NOT_RUNNING"
	AppStatusRunning    = "RUNNING"
	AppStatusFrozen     = "FROZEN"
)

type PrivateNetworkAccess struct {
	ApnUUID              string      `json:"apnUUID,omitempty"`
	Networks             []Network   `json:"networks,omitempty"`
	CustomerNetwork      string      `json:"customerNetwork"`
	AdditionalRoutes     []string    `json:"additionalRoutes,omitempty"`
	NetworkInterfaceName string      `json:"networkInterfaceName,omitempty"`
	AppPodFixIp          AppPodFixIp `json:"appPodFixIp,omitempty"`
}

type AppPodFixIp struct {
	Db string `json:"db"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulSpec defines the desired state of Consul
type ConsulSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ReplicaCount         int                   `json:"replicaCount"`
	Ports                Ports                 `json:"ports"`
	MetricsDomainName    string                `json:"metricsDomainName,omitempty"`
	PrivateNetworkAccess *PrivateNetworkAccess `json:"privateNetworkAccess,omitempty"`
}

type AppReporteData struct {
	//The structure of this type is up the application. AppFw will convert the whole representation to JSON.
	MetricsClusterIp string `json:"metricsClusterIp,omitempty"`
	//Ip addresses of the services that received IP address from the private network
	PrivateNetworkIpAddress map[string]string `json:"privateNetworkIpAddresses,omitempty"`
}

// ConsulStatus defines the observed state of Consul
type ConsulStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
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

// +kubebuilder:object:root=true

// ConsulList contains a list of Consul
type ConsulList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Consul `json:"items"`
}

type Network struct {
	ApnUUID          string   `json:"apnUUID,omitempty"`
	NetworkID        string   `json:"networkId,omitempty"`
	AdditionalRoutes []string `json:"additionalRoutes,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Consul{}, &ConsulList{})
}
