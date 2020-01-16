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

// ShutterSpec defines the desired state of Shutter
type ShutterSpec struct {
	ClosedPercentage int `json:"closedPercentage"`
}

type ShutterPhaseTypes string

const (
	ShutterMoving = "Moving"
	ShutterIdle   = "Idle"
)

// ShutterStatus defines the observed state of Shutter
type ShutterStatus struct {
	ObservedGeneration int64             `json:"observedGeneration,omitempty"`
	Phase              ShutterPhaseTypes `json:"phase,omitempty"`
	ClosedPercentage   int               `json:"closedPercentage"`
}

// Shutter is the Schema for the shutters API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Target",type="string",JSONPath=".spec.closedPercentage"
// +kubebuilder:printcolumn:name="Current",type="string",JSONPath=".status.closedPercentage"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Shutter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShutterSpec   `json:"spec,omitempty"`
	Status ShutterStatus `json:"status,omitempty"`
}

// ShutterList contains a list of Shutter
// +kubebuilder:object:root=true
type ShutterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Shutter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Shutter{}, &ShutterList{})
}
