package duck

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/apis/duck"
)

type UntypedInformerFactory struct {
	Client       dynamic.Interface
	ResyncPeriod time.Duration
	StopChannel  <-chan struct{}
}

var _ duck.InformerFactory = (*UntypedInformerFactory)(nil)

func (dif *UntypedInformerFactory) Get(gvr schema.GroupVersionResource) (cache.SharedIndexInformer, cache.GenericLister, error) {
	// Avoid error cases, like the GVR does not exist.
	// It is not a full check. Some RBACs might sneak by, but the window is very small.
	if _, err := dif.Client.Resource(gvr).List(metav1.ListOptions{}); err != nil {
		return nil, nil, err
	}

	lw := &cache.ListWatch{
		ListFunc:  asUnstructuredList(dif.Client.Resource(gvr).List),
		WatchFunc: dif.Client.Resource(gvr).Watch,
	}
	inf := cache.NewSharedIndexInformer(lw, &unstructured.Unstructured{}, dif.ResyncPeriod, cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})

	lister := cache.NewGenericLister(inf.GetIndexer(), gvr.GroupResource())

	go inf.Run(dif.StopChannel)

	if ok := cache.WaitForCacheSync(dif.StopChannel, inf.HasSynced); !ok {
		return nil, nil, fmt.Errorf("failed starting shared index informer for %v with type Unstructured", gvr)
	}

	return inf, lister, nil
}

type unstructuredList func(opts metav1.ListOptions) (*unstructured.UnstructuredList, error)

func asUnstructuredList(l unstructuredList) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		return l(opts)
	}
}
