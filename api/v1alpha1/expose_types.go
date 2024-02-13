/*
Copyright 2023 freedisch.

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

// ExposeSpec defines the desired state of Expose
type ExposeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Expose. Edit expose_types.go to remove/update
	Name       string           `json:"name,omitempty"`
	Deployment []DeploymentSpec `json:"deployment"`
	Service    []ServiceSpec    `json:"service"`
	Ingress    []IngressSpec    `json:"ingress"`
}

type DeploymentSpec struct {
	Name       string          `json:"name,omitempty"`
	Replicas   int32           `json:"replicas,omitempty"`
	Component  string          `json:"component,omitempty"`
	Containers []ContainerSpec `json:"containers"`
}

type ServiceSpec struct {
	Name string `json:"name,omitempty"`
	Port int32  `json:"port"`
}

type IngressSpec struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

type ContainerSpec struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

// ExposeStatus defines the observed state of Expose
type ExposeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Expose is the Schema for the exposes API
type Expose struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExposeSpec   `json:"spec,omitempty"`
	Status ExposeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExposeList contains a list of Expose
type ExposeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Expose `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Expose{}, &ExposeList{})
}
