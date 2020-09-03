package v1alpha1

import (
	"context"
)

// SetDefaults implements apis.Defaultable
func (b *Binding) SetDefaults(ctx context.Context) {
	if b.Spec.Subject.Namespace == "" {
		b.Spec.Subject.Namespace = b.Namespace
	}
	for i := range b.Spec.Backings {
		if b.Spec.Backings[i].Namespace == "" {
			b.Spec.Backings[i].Namespace = b.Namespace
		}
	}
}
