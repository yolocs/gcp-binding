package v1alpha1

import (
	"context"

	"go.uber.org/zap"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"

	untypedmutations "github.com/yolocs/gcp-binding/pkg/mutations/untyped"

	_ "github.com/yolocs/gcp-binding/pkg/mutations/untyped/sql"
)

func (b *Binding) Do(ctx context.Context, ps *duckv1.WithPod) {
	b.Undo(ctx, ps)

	backings := GetBackings(ctx)

	for _, b := range backings {
		m, ok := untypedmutations.Default[b.GroupVersionKind()]
		if !ok {
			logging.FromContext(ctx).Desugar().Warn("Backing ref doesn't have corresponding mutation", zap.Any("backing", b.GroupVersionKind()))
			return
		}
		if err := m.Do(ctx, ps, b); err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed mutation for backing ref", zap.Any("backing", b.GroupVersionKind()))
		}
	}
}

func (b *Binding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	backings := GetBackings(ctx)

	for _, b := range backings {
		m, ok := untypedmutations.Default[b.GroupVersionKind()]
		if !ok {
			logging.FromContext(ctx).Desugar().Warn("Backing ref doesn't have corresponding mutation", zap.Any("backing", b.GroupVersionKind()))
			return
		}
		if err := m.Undo(ctx, ps, b); err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed mutation for backing ref", zap.Any("backing", b.GroupVersionKind()))
		}
	}
}
