package untyped

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	pkgapisduck "knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/client/injection/ducks/duck/v1/conditions"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/tracker"
)

type BackingResolver struct {
	tracker         tracker.Interface
	informerFactory pkgapisduck.InformerFactory
}

func NewBackingResolver(ctx context.Context, callback func(types.NamespacedName)) *BackingResolver {
	ret := &BackingResolver{}

	ret.tracker = tracker.New(callback, controller.GetTrackerLease(ctx))
	ret.informerFactory = &pkgapisduck.CachedInformerFactory{
		Delegate: &pkgapisduck.EnqueueInformerFactory{
			Delegate:     conditions.Get(ctx),
			EventHandler: controller.HandleAll(ret.tracker.OnChanged),
		},
	}

	return ret
}

func (r *BackingResolver) ResolveBackingFromRef(kref duckv1.KReference, parent interface{}) (*unstructured.Unstructured, error) {
	ref := corev1.ObjectReference{Name: kref.Name, Namespace: kref.Namespace, APIVersion: kref.APIVersion, Kind: kref.Kind}
	if err := r.tracker.TrackReference(tracker.Reference{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Name:       ref.Name,
		Namespace:  ref.Namespace,
	}, parent); err != nil {
		return nil, fmt.Errorf("failed to track backing ref %+v: %w", ref, err)
	}
	return r.getBacking(ref)
}

func (r *BackingResolver) getBacking(ref corev1.ObjectReference) (*unstructured.Unstructured, error) {
	gvr, _ := meta.UnsafeGuessKindToResource(ref.GroupVersionKind())
	_, lister, err := r.informerFactory.Get(gvr)
	if err != nil {
		return nil, fmt.Errorf("failed to get lister for %+v: %w", gvr, err)
	}

	obj, err := lister.ByNamespace(ref.Namespace).Get(ref.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get ref %+v: %w", ref, err)
	}

	return obj.(*unstructured.Unstructured), nil
}
