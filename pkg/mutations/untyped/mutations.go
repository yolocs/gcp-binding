package untyped

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var Default = map[schema.GroupVersionKind]Mutation{}

type Mutation interface {
	Do(context.Context, *duckv1.WithPod, *unstructured.Unstructured) error
	Undo(context.Context, *duckv1.WithPod, *unstructured.Unstructured) error
}
