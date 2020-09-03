package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolocs/gcp-binding/pkg/mutations/untyped"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func init() {
	untyped.Default[schema.GroupVersionKind{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLDatabase"}] = forDatabase{}
	untyped.Default[schema.GroupVersionKind{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLUser"}] = forUser{}
	untyped.Default[schema.GroupVersionKind{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLInstance"}] = forInstance{}
}

type forInstance struct{}

func (m forInstance) Do(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	conn, err := getInstanceConnName(u)
	if err != nil {
		return err
	}
	spec := ps.Spec.Template.Spec
	spec.Containers = append(spec.Containers, corev1.Container{
		Name:  "cloud-sql-proxy",
		Image: "gcr.io/cloudsql-docker/gce-proxy:1.17",
		Command: []string{
			"/cloud_sql_proxy",
			fmt.Sprintf("-instances=%s=tcp:3306", conn),
		},
		// SecurityContext: &corev1.SecurityContext{RunAsNonRoot: ptr.Bool(true)},
	})

	ps.Spec.Template.Spec.Containers = spec.Containers

	return nil
}

func (m forInstance) Undo(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	cs := []corev1.Container{}
	for _, c := range ps.Spec.Template.Spec.Containers {
		if c.Name != "cloud-sql-proxy" {
			cs = append(cs, c)
		}
	}
	ps.Spec.Template.Spec.Containers = cs
	return nil
}

type forUser struct{}

func (m forUser) Do(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	un := u.GetName()
	upass, err := getPass(u)
	if err != nil {
		return err
	}
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		if c.Name != "cloud-sql-proxy" {
			spec.Containers[i].Env = append(
				spec.Containers[i].Env,
				corev1.EnvVar{
					Name:  "DB_USER",
					Value: un,
				}, corev1.EnvVar{
					Name:      "DB_PASS",
					ValueFrom: upass,
				})
		}
	}
	return nil
}

func (m forUser) Undo(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		if c.Name != "cloud-sql-proxy" {
			envs := []corev1.EnvVar{}
			for _, e := range c.Env {
				if e.Name != "DB_USER" && e.Name != "DB_PASS" {
					envs = append(envs, e)
				}
			}
			spec.Containers[i].Env = envs
		}
	}
	return nil
}

type forDatabase struct{}

func (m forDatabase) Do(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	db := u.GetName()
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		if c.Name != "cloud-sql-proxy" {
			spec.Containers[i].Env = append(
				spec.Containers[i].Env,
				corev1.EnvVar{
					Name:  "DB_NAME",
					Value: db,
				})
		}
	}
	return nil
}

func (m forDatabase) Undo(ctx context.Context, ps *duckv1.WithPod, u *unstructured.Unstructured) error {
	spec := ps.Spec.Template.Spec
	for i, c := range spec.Containers {
		if c.Name != "cloud-sql-proxy" {
			envs := []corev1.EnvVar{}
			for _, e := range c.Env {
				if e.Name != "DB_NAME" {
					envs = append(envs, e)
				}
			}
			spec.Containers[i].Env = envs
		}
	}
	return nil
}

func getPass(u *unstructured.Unstructured) (*corev1.EnvVarSource, error) {
	spec := u.Object["spec"].(map[string]interface{})
	passRaw, ok := spec["password"]
	if !ok {
		return nil, errors.New("SQLUser password ref not found in spec")
	}
	pass, ok := passRaw.(map[string]interface{})
	if !ok || pass == nil {
		return nil, errors.New("SQLUser password ref not found in spec")
	}
	passRef := pass["valueFrom"].(map[string]interface{})["secretKeyRef"].(map[string]interface{})
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: passRef["name"].(string)},
			Key:                  passRef["key"].(string),
		},
	}, nil
}

func getInstanceConnName(u *unstructured.Unstructured) (string, error) {
	status, ok := u.Object["status"].(map[string]interface{})
	if !ok {
		return "", errors.New("SQLInstance connection name not found in status")
	}
	connRaw, ok := status["connectionName"]
	if !ok {
		return "", errors.New("SQLInstance connection name not found in status")
	}
	conn, ok := connRaw.(string)
	if !ok {
		return "", errors.New("SQLInstance connection name not found in status")
	}
	return conn, nil
}
