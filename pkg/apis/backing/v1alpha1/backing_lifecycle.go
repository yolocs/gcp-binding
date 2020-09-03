package v1alpha1

import (
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/tracker"
)

const (
	BindingConditionReady = apis.ConditionReady
	// BindingConditionBackings apis.ConditionType = "BackingsResolved"
)

var bindingCondSet = apis.NewLivingConditionSet(
// BindingConditionBackings,
)

func (b *Binding) GetConditionSet() apis.ConditionSet {
	return bindingCondSet
}

func (bs *BindingStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return bindingCondSet.Manage(bs).GetCondition(t)
}

func (bs *BindingStatus) IsReady() bool {
	return bindingCondSet.Manage(bs).IsHappy()
}

func (bs *BindingStatus) InitializeConditions() {
	bindingCondSet.Manage(bs).InitializeConditions()
}

func (bs *BindingStatus) SetObservedGeneration(gen int64) {
	bs.ObservedGeneration = gen
}

// func (bs *BindingStatus) MarkBackingsResolvedSuccess() {
// 	bindingCondSet.Manage(bs).MarkTrue(BindingConditionBackings)
// }

// func (bs *BindingStatus) MarkBackingsResolvedFailure(reason, messageFormat string, messageA ...interface{}) {
// 	bindingCondSet.Manage(bs).MarkFalse(BindingConditionBackings, reason, messageFormat, messageA...)
// }

func (bs *BindingStatus) MarkBindingAvailable() {
	bindingCondSet.Manage(bs).MarkTrue(BindingConditionReady)
}

func (bs *BindingStatus) MarkBindingUnavailable(reason, message string) {
	bindingCondSet.Manage(bs).MarkFalse(BindingConditionReady, reason, message)
}

func (b *Binding) GetBindingStatus() duck.BindableStatus {
	return &b.Status
}

// GetSubject implements psbinding.Bindable
func (b *Binding) GetSubject() tracker.Reference {
	return b.Spec.Subject
}
