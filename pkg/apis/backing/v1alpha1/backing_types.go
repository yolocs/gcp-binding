package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Binding is to bind a subject to GCP backing services.
type Binding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec"`
	Status BindingStatus `json:"status"`
}

var (
	_ runtime.Object     = (*Binding)(nil)
	_ kmeta.OwnerRefable = (*Binding)(nil)
	_ apis.Validatable   = (*Binding)(nil)
	_ apis.Defaultable   = (*Binding)(nil)
	_ apis.HasSpec       = (*Binding)(nil)
	_ duckv1.KRShaped    = (*Binding)(nil)
)

// BindingSpec is the binding spec.
type BindingSpec struct {
	// inherits duck/v1beta1 BindingSpec, which currently provides:
	// * Subject - Subject references the resource(s) whose "runtime contract"
	//   should be augmented by Binding implementations.
	duckv1beta1.BindingSpec `json:",inline"`

	// A list of backing services to bind.
	Backings []duckv1.KReference `json:"backings"`
}

// BindingStatus is the binding status.
type BindingStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BindingList is a collection of Binding.
type BindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Binding `json:"items"`
}

// GetStatus retrieves the status of the SinkBinding. Implements the KRShaped interface.
func (s *Binding) GetStatus() *duckv1.Status {
	return &s.Status.Status
}

// GetGroupVersionKind returns GroupVersionKind for EventPolicy
func (b *Binding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Binding")
}

// GetUntypedSpec returns the spec of the EventPolicy.
func (b *Binding) GetUntypedSpec() interface{} {
	return b.Spec
}
