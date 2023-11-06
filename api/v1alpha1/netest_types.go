/*
Copyright 2023.
*/

package v1alpha1

import (
	"github.com/terloo/kubenetest-operator/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetestSpec defines the desired state of Netest
type NetestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

// NetestStatus defines the observed state of Netest
type NetestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TestItems map[meta.NetestType]*NetestResult `json:"testItems,omitempty"`
}

type NetestResultState int

const (
	Queue     NetestResultState = iota + 1
	Runing    NetestResultState = iota + 1
	Completed NetestResultState = iota + 1
)

type NetestResult struct {
	Type   meta.NetestType   `json:"type,omitempty"`
	Status NetestResultState `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+genclient

// Netest is the Schema for the netests API
type Netest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetestSpec   `json:"spec,omitempty"`
	Status NetestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetestList contains a list of Netest
type NetestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Netest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Netest{}, &NetestList{})
}
