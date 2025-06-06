/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Evpn4OvnSpec defines the desired state of Evpn4Ovn
type Evpn4OvnSpec struct {
	ConfigMap       string         `json:"configMap"`
	Replicas        int32          `json:"replicas,omitempty"`
	Api             ApiSpec        `json:"api"`
	MpbgpAgent      MpbgpAgentSpec `json:"mpbgpAgent,omitempty"`
	OvsdbAgent      OvsdbAgentSpec `json:"ovsdbAgent,omitempty"`
	RabbitmqService ServiceSpec    `json:"rabbitmqService,omitempty"`
	BgpService      ServiceSpec    `json:"bgpService,omitempty"`
	OvsdbService    ServiceSpec    `json:"ovsdbService,omitempty"`
}

// Evpn4OvnSpec defines the desired state of Evpn4Ovn
type ApiSpec struct {
	//Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
	Port  int32  `json:"port,omitempty"`
}

type MpbgpAgentSpec struct {
	//Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

type OvsdbAgentSpec struct {
	//Name		string `json:"name,omitempty"`
	Image       string `json:"image,omitempty"`
	OVNProvider string `json:"ovnProvider,omitempty"`
}

type ServiceSpec struct {
	Name      string        `json:"name,omitempty"`
	Endpoints EndpointsSpec `json:"endpoints,omitempty"`
}

type EndpointsSpec struct {
	//Name		string `json:"name,omitempty"`
	IP    []string   `json:"ip,omitempty"`
	Ports []PortSpec `json:"ports,omitempty"`
}

type PortSpec struct {
	Name string `json:"name,omitempty"`
	Port int32  `json:"port,omitempty"`
}

// Evpn4OvnStatus defines the observed state of Evpn4Ovn
type Evpn4OvnStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Evpn4Ovn is the Schema for the evpn4ovns API
type Evpn4Ovn struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Evpn4OvnSpec   `json:"spec,omitempty"`
	Status Evpn4OvnStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// Evpn4OvnList contains a list of Evpn4Ovn
type Evpn4OvnList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Evpn4Ovn `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Evpn4Ovn{}, &Evpn4OvnList{})
}
