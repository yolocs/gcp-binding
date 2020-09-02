package v1alpha1

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
)

func (b *Binding) Do(ctx context.Context, ps *duckv1.WithPod) {
	// TODO
	b.Undo(ctx, ps)

	backings := GetBackings(ctx)
	logging.FromContext(ctx).Desugar().Info("Backings CSHOU DEBUG", zap.Any("Backings", backings))

	spec := ps.Spec.Template.Spec
	for i := range spec.Containers {
		spec.Containers[i].Env = append(spec.Containers[i].Env, corev1.EnvVar{
			Name:  "K_TEST",
			Value: "foobar",
		})
	}
}

func (b *Binding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	// TODO
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		if len(c.Env) == 0 {
			continue
		}
		env := make([]corev1.EnvVar, 0, len(spec.Containers[i].Env))
		for j, ev := range c.Env {
			switch ev.Name {
			case "K_TEST":
				continue
			default:
				env = append(env, spec.Containers[i].Env[j])
			}
		}
		spec.Containers[i].Env = env
	}
}
