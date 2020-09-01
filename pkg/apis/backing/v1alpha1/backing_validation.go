package v1alpha1

import (
	"context"
	"errors"
	"strings"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// Validate implements apis.Validatable
func (b *Binding) Validate(ctx context.Context) *apis.FieldError {
	withNS := apis.WithinParent(ctx, b.ObjectMeta)
	return b.Spec.Validate(withNS).ViaField("spec")
}

// Validate implements apis.Validatable
func (bs *BindingSpec) Validate(ctx context.Context) *apis.FieldError {
	errs := bs.Subject.Validate(ctx).ViaField("subject")
	if bs.Subject.Namespace != apis.ParentMeta(ctx).Namespace {
		errs = errs.Also(apis.ErrInvalidValue(bs.Subject.Namespace, "subject.namespace"))
	}
	if len(bs.Backings) == 0 {
		errs = errs.Also(apis.ErrInvalidValue(bs.Backings, "backings"))
	}
	for i, backing := range bs.Backings {
		if err := validateBacking(ctx, backing); err != nil {
			errs = errs.Also(apis.ErrInvalidArrayValue(backing, "backings", i))
		}
	}
	return errs
}

func validateBacking(ctx context.Context, backing duckv1.KReference) error {
	parts := strings.Split(backing.APIVersion, "/")
	if !strings.HasSuffix(parts[0], ".cnrm.cloud.google.com") {
		return errors.New("Invalid backing service ref API group")
	}
	if backing.Namespace != apis.ParentMeta(ctx).Namespace {
		return errors.New("Invalid backing service ref namespace")
	}
	return nil
}
