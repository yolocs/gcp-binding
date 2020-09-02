package backing

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podspecable"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/namespace"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"

	"github.com/yolocs/gcp-binding/pkg/apis/backing/v1alpha1"
	bbinformer "github.com/yolocs/gcp-binding/pkg/client/injection/informers/backing/v1alpha1/binding"
	"github.com/yolocs/gcp-binding/pkg/resolver/untyped"
)

const (
	controllerAgentName = "backingbinding-controller"
)

func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	bbInformer := bbinformer.Get(ctx)
	dc := dynamicclient.Get(ctx)
	psInformerFactory := podspecable.Get(ctx)
	namespaceInformer := namespace.Get(ctx)

	c := &psbinding.BaseReconciler{
		LeaderAwareFuncs: reconciler.LeaderAwareFuncs{
			PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
				all, err := bbInformer.Lister().List(labels.Everything())
				if err != nil {
					return err
				}
				for _, elt := range all {
					enq(bkt, types.NamespacedName{
						Namespace: elt.GetNamespace(),
						Name:      elt.GetName(),
					})
				}
				return nil
			},
		},
		GVR: v1alpha1.SchemeGroupVersion.WithResource("bindings"),
		Get: func(namespace string, name string) (psbinding.Bindable, error) {
			return bbInformer.Lister().Bindings(namespace).Get(name)
		},
		DynamicClient: dc,
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
		NamespaceLister: namespaceInformer.Lister(),
	}
	impl := controller.NewImpl(c, logger, "BackingBindings")

	logger.Info("Setting up event handlers")

	bbInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	namespaceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	c.WithContext = WithContextFactory(ctx, impl.EnqueueKey)
	c.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	c.Factory = &duck.CachedInformerFactory{
		Delegate: &duck.EnqueueInformerFactory{
			Delegate:     psInformerFactory,
			EventHandler: controller.HandleAll(c.Tracker.OnChanged),
		},
	}

	return impl
}

func ListAll(ctx context.Context, handler cache.ResourceEventHandler) psbinding.ListAll {
	bbInformer := bbinformer.Get(ctx)

	// Whenever a Binding changes our webhook programming might change.
	bbInformer.Informer().AddEventHandler(handler)

	return func() ([]psbinding.Bindable, error) {
		l, err := bbInformer.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
		bl := make([]psbinding.Bindable, 0, len(l))
		for _, elt := range l {
			bl = append(bl, elt)
		}
		return bl, nil
	}

}

func WithContextFactory(ctx context.Context, handler func(types.NamespacedName)) psbinding.BindableContext {
	r := untyped.NewBackingResolver(ctx, handler)

	return func(ctx context.Context, b psbinding.Bindable) (context.Context, error) {
		bb := b.(*v1alpha1.Binding)
		backings := make([]*unstructured.Unstructured, 0)
		for _, ref := range bb.Spec.Backings {
			backing, err := r.ResolveBackingFromRef(ref, bb)
			if err != nil {
				return nil, err
			}
			backings = append(backings, backing)
		}
		return v1alpha1.WithBackings(ctx, backings), nil
	}
}
