package matcher

import (
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
	"time"
)

type K8sExistsMatcher struct {
	Timeout time.Duration
}

func (k K8sExistsMatcher) Match(actual interface{}) (success bool, err error) {
	actualTyped, ok := actual.(K8sResourceId)
	if !ok {
		return false, errors.New("actual param is not a K8sResourceId type")
	}

	gvr, apiResource, err := GetGvrAndAPIResources(actualTyped.Gvk)
	if err != nil {
		return false, errors.Wrap(err, "failed to get the GVR of the given resource")
	}

	_, err = GetCurrentStateOfResource(apiResource, gvr, actualTyped.Name, actualTyped.Namespace)

	if err == nil {
		return true, nil
	} else if k.Timeout == 0 {
		return false, err
	}

	stopper := make(chan struct{})

	//sometimes watcher doesn't stop the watching after closing the stopper channel and
	//other events arrive which causes panic because we would like to close a closed channel
	//that's why the purpose of this flag is to indicate when the channel is closed once
	finished := false
	result := false
	err = watchWithTimeout(
		actualTyped.Name,
		actualTyped.Namespace,
		"",
		gvr,
		stopper,
		k.Timeout,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if !finished {
					FwLog.V(0).Info("ExitsMatcher, Add event", "obj", obj)
					result = true
					close(stopper)
					finished = true
				}
			},
			DeleteFunc: func(obj interface{}) {
				if !finished {
					FwLog.V(0).Info("EqualsMatcher, Delete event", "obj", obj)
					close(stopper)
					result = false
					finished = true
				}
			},
		},
	)
	if err != nil {
		FwLog.Info("Watch stopped", "err", err)
		return false, nil
	}
	return result, nil
}

func (k K8sExistsMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to exists")
}

func (k K8sExistsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to exists")
}

func ExistsK8sRes(waitTimeout ...time.Duration) types.GomegaMatcher {
	var usedTimeout time.Duration

	if len(waitTimeout) > 0 {
		usedTimeout = waitTimeout[0]
	} else {
		usedTimeout = 0
	}

	return K8sExistsMatcher{Timeout: usedTimeout}
}
