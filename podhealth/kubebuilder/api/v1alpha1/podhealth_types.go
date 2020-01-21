/*

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

// PodHealthSpec defines the desired state of PodHealth
type PodHealthSpec struct {
	// PodSelector selects pods to get the Health for
	PodSelector metav1.LabelSelector `json:"podSelector,omitempty"`
}

// PodHealthStatus defines the observed state of PodHealth
type PodHealthStatus struct {
	Ready       int         `json:"ready,omitempty"`
	Unready     int         `json:"unready,omitempty"`
	Total       int         `json:"total,omitempty"`
	LastChecked metav1.Time `json:"lastChecked,omitempty"`
}

// PodHealth is the Schema for the podhealths API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Unready",type="integer",JSONPath=".status.unready"
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".status.total"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=ph
type PodHealth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodHealthSpec   `json:"spec,omitempty"`
	Status PodHealthStatus `json:"status,omitempty"`
}

// PodHealthList contains a list of PodHealth
// +kubebuilder:object:root=true
type PodHealthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodHealth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodHealth{}, &PodHealthList{})
}
