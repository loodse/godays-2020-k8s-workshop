package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodHealthSpec defines the desired state of PodHealth
// +k8s:openapi-gen=true
type PodHealthSpec struct {
	// PodSelector selects pods to get the Health for
	PodSelector metav1.LabelSelector `json:"podSelector,omitempty"`
}

// PodHealthStatus defines the observed state of PodHealth
// +k8s:openapi-gen=true
type PodHealthStatus struct {
	Ready       int         `json:"ready,omitempty"`
	Unready     int         `json:"unready,omitempty"`
	Total       int         `json:"total,omitempty"`
	LastChecked metav1.Time `json:"lastChecked,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodHealth is the Schema for the podhealths API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=podhealths,scope=Namespaced
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodHealthList contains a list of PodHealth
type PodHealthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodHealth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodHealth{}, &PodHealthList{})
}
