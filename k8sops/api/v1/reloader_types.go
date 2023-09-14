/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReloaderSpec defines the desired state of Reloader
type ReloaderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Reloader. Edit reloader_types.go to remove/update
	ConfigMap string            `json:"configmap,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
}

// ReloaderStatus defines the observed state of Reloader
type ReloaderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Type string `json:"type,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Reloader is the Schema for the reloaders API
type Reloader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReloaderSpec   `json:"spec,omitempty"`
	Status ReloaderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReloaderList contains a list of Reloader
type ReloaderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Reloader `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Reloader{}, &ReloaderList{})
}
