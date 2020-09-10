package storage

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/yolocs/gcp-binding/pkg/mutations/untyped"
)

func init() {
	untyped.Default[schema.GroupVersionKind{Group: "storage.cnrm.cloud.google.com", Version: "v1beta1", Kind: "StorageBucket"}] = forBucket{}
}

type forBucket struct{}

func (m forBucket) Do(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	spec := ps.Spec.Template.Spec
	for i := range spec.Containers {
		spec.Containers[i].Env = append(
			spec.Containers[i].Env,
			corev1.EnvVar{
				Name:  "BUCKET_NAME",
				Value: u.GetName(),
			})
	}
	return nil
}

func (m forBucket) Undo(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		envs := []corev1.EnvVar{}
		for _, e := range c.Env {
			if e.Name != "BUCKET_NAME" {
				envs = append(envs, e)
			}
		}
		spec.Containers[i].Env = envs
	}
	return nil
}
