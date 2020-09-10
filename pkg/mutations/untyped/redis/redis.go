package redis

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/yolocs/gcp-binding/pkg/mutations/untyped"
)

func init() {
	untyped.Default[schema.GroupVersionKind{Group: "redis.cnrm.cloud.google.com", Version: "v1beta1", Kind: "RedisInstance"}] = forInstance{}
}

type forInstance struct{}

func (m forInstance) Do(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	hostIP, err := getSubVal(u, "status", "host")
	if err != nil {
		return err
	}
	port, err := getSubVal(u, "status", "port")
	if err != nil {
		return err
	}
	reservedIPRange, err := getSubVal(u, "spec", "reservedIpRange")
	if err != nil {
		return err
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.Containers {
		spec.Containers[i].Env = append(
			spec.Containers[i].Env,
			corev1.EnvVar{
				Name:  "REDIS_HOST",
				Value: hostIP,
			}, corev1.EnvVar{
				Name:  "REDIS_PORT",
				Value: port,
			}, corev1.EnvVar{
				Name:  "REDIS_IP_RANGE",
				Value: reservedIPRange,
			})
	}
	return nil
}

func (m forInstance) Undo(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		envs := []corev1.EnvVar{}
		for _, e := range c.Env {
			if e.Name != "REDIS_HOST" && e.Name != "REDIS_PORT" && e.Name != "REDIS_IP_RANGE" {
				envs = append(envs, e)
			}
		}
		spec.Containers[i].Env = envs
	}
	return nil
}

func getSubVal(u *unstructured.Unstructured, section, key string) (string, error) {
	status, ok := u.Object[section].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("RedisInstance %q not found in status", key)
	}
	valRaw, ok := status[key]
	if !ok {
		return "", fmt.Errorf("RedisInstance %q not found in status", key)
	}
	val, ok := valRaw.(string)
	if !ok {
		return "", fmt.Errorf("RedisInstance %q not found in status", key)
	}
	return val, nil
}
