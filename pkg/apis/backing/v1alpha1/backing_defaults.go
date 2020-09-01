package v1alpha1

import (
	"context"
)

// SetDefaults implements apis.Defaultable
func (b *Binding) SetDefaults(ctx context.Context) {
	if b.Spec.Subject.Namespace == "" {
		b.Spec.Subject.Namespace = b.Namespace
	}
	for _, backing := range b.Spec.Backings {
		if backing.Namespace == "" {
			backing.Namespace = b.Namespace
		}
	}
}
