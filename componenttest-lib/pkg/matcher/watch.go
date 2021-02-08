package matcher

import (
	"github.com/pkg/errors"
	"gitlabe2.ext.net.nokia.com/Nokia_DAaaS/edge-microservices/componenttest-lib/pkg/env"
	"gitlabe2.ext.net.nokia.com/Nokia_DAaaS/edge-microservices/componenttest-lib/pkg/k8sclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"time"
)

func watchWithTimeout(resName, resNamespace string, resourceVersion string, gvr schema.GroupVersionResource, stopper chan struct{}, timeout time.Duration, eventHandle cache.ResourceEventHandler) error {
	go watchInformer(resName, resNamespace, resourceVersion, gvr, stopper, eventHandle)

	if waitTimeout(stopper, timeout) {
		close(stopper)
		return errors.New("watching has been timed out")
	}
	return nil
}

func waitTimeout(finished <-chan struct{}, timeout time.Duration) bool {
	select {
	case <-finished:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func watchInformer(name string, namespace string, resourceVersion string, gvr schema.GroupVersionResource, stopper <-chan struct{}, eventHandle cache.ResourceEventHandler) {

	listOp := getListOptions(name, resourceVersion)
	listOptionsFunc := dynamicinformer.TweakListOptionsFunc(func(options *v1.ListOptions) { *options = listOp })

	dynInformer := dynamicinformer.NewFilteredDynamicInformer(
		k8sclient.GetDynamicK8sClient(env.Cfg),
		gvr,
		namespace,
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		listOptionsFunc)

	informer := dynInformer.Informer()
	informer.AddEventHandler(eventHandle)

	informer.Run(stopper)
}

func getListOptions(name string, resourceVersion string) v1.ListOptions {
	listOp := v1.ListOptions{
		FieldSelector: "metadata.name=" + name,
	}
	if resourceVersion != "" {
		listOp.ResourceVersion = resourceVersion
	}

	return listOp
}

func GetAPIResourceByGvk(gvk schema.GroupVersionKind) (v1.APIResource, error) {
	if gvk.Version == "" || gvk.Kind == "" {
		return v1.APIResource{}, errors.New("empty input parameters")
	}

	groupVersion := ""
	if gvk.Group == "" {
		groupVersion = gvk.Version
	} else {
		groupVersion = gvk.Group + "/" + gvk.Version
	}
	resList, err := k8sclient.GetK8sClient(env.Cfg).Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return v1.APIResource{}, err
	}

	for _, res := range resList.APIResources {
		if res.Kind == gvk.Kind {
			return res, nil
		}
	}

	return v1.APIResource{}, errors.New("not found")
}
