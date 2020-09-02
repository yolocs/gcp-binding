package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type backingsKey struct{}

func WithBackings(ctx context.Context, backings []*unstructured.Unstructured) context.Context {
	return context.WithValue(ctx, backingsKey{}, backings)
}

func GetBackings(ctx context.Context) []*unstructured.Unstructured {
	value := ctx.Value(backingsKey{})
	if value == nil {
		return nil
	}
	return value.([]*unstructured.Unstructured)
}
