/*
Copyright 2020 Nokia
Licensed under the BSD 3-Clause License.
SPDX-License-Identifier: BSD-3-Clause
*/

package matcher

import (
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"time"
)

const DefaultWaitTimeout = 10 * time.Second
var FwLog = logf.Log.WithName("matcher")

type K8sEqualsMatcher struct {
	Expected interface{}
	Timeout  time.Duration
}

func (m K8sEqualsMatcher) Match(actual interface{}) (success bool, err error) {
	actualTyped, ok := actual.(K8sResourceId)
	if !ok {
		return false, errors.New("actual param is not a K8sResourceParamId type")
	}

	gvr, apiResource, err := GetGvrAndAPIResources(actualTyped.Gvk)
	if err != nil {
		return false, errors.Wrap(err, "failed to get the GVR of the given resource")
	}

	actInstance, err := GetCurrentStateOfResource(apiResource, gvr, actualTyped.Name, actualTyped.Namespace)
	resourceVersion := ""
	if err == nil {
		if m.IsMatch(actInstance, actualTyped.ParamPath) {
			return true, nil
		}
		if m.isWatchDisabled() {
			return false, nil
		}
		resourceVersion = actInstance.GetResourceVersion()
	}

	stopper := make(chan struct{})

	result := false
	//sometimes watcher doesn't stop the watching after closing the stopper channel and
	//other events arrive which causes panic because we would like to close a closed channel
	//that's why the purpose of this flag is to indicate when the channel is closed once
	finished := false
	err = watchWithTimeout(
		actualTyped.Name,
		actualTyped.Namespace,
		resourceVersion,
		gvr,
		stopper,
		m.Timeout,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if !finished {
					FwLog.V(0).Info("EqualsMatcher, Delete event", "obj", obj)
					close(stopper)
					finished = true
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if !finished && m.IsMatch(newObj, actualTyped.ParamPath) {
					FwLog.V(0).Info("EqualsMatcher, Update event", "oldObj", oldObj, "newObj", newObj)
					result = true
					finished = true
					close(stopper)
				}
			},
			AddFunc: func(obj interface{}) {
				if !finished && m.IsMatch(obj, actualTyped.ParamPath) {
					FwLog.V(0).Info("EqualsMatcher, Add event", "obj", obj)
					result = true
					finished = true
					close(stopper)
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

func (m K8sEqualsMatcher) isWatchDisabled() bool {
	return m.Timeout == 0
}

func (m K8sEqualsMatcher) IsMatch(obj interface{}, path []string) bool {
	if field, found := getFieldByPath(obj, path); found {
		return reflect.DeepEqual(field, m.Expected)
	}
	return false
}

func (m K8sEqualsMatcher) FailureMessage(actual interface{}) (message string) {
	if field, found := getFieldByActual(actual); found {
		if _, ok := field.(string); ok {
			return format.MessageWithDiff(field.(string), "to equal", m.Expected.(string))
		}
		return format.Message(field, "to equal", m.Expected)
	}
	return format.Message(actual, "(doesn't exist)\nto equal", m.Expected)
}

func (m K8sEqualsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if field, found := getFieldByActual(actual); found {
		if _, ok := field.(string); ok {
			return format.MessageWithDiff(field.(string), "not to equal", m.Expected.(string))
		}
		return format.Message(field, "not to equal", m.Expected)
	}
	return format.Message(actual, "(doesn't exist)\nnot to equal", m.Expected)
}

func EqualsK8sRes(expected interface{}, timeout ...time.Duration) types.GomegaMatcher {
	var usedTimeout time.Duration

	if len(timeout) > 0 {
		usedTimeout = timeout[0]
	} else {
		usedTimeout = DefaultWaitTimeout
	}
	return K8sEqualsMatcher{Expected: expected, Timeout: usedTimeout}
}

func getFieldByActual(actual interface{}) (interface{}, bool) {
	actualTyped, ok := actual.(K8sResourceId)
	if !ok {
		return nil, false
	}

	actInstance, err := getInstanceByActual(actualTyped)
	if err != nil {
		return nil, false
	}

	field, found := getFieldByPath(actInstance, actualTyped.ParamPath)
	return field, found
}

func getInstanceByActual(actualTyped K8sResourceId) (*unstructured.Unstructured, error) {
	gvr, apiResource, err := GetGvrAndAPIResources(actualTyped.Gvk)
	if err != nil {
		return nil, nil
	}

	actInstance, err := GetCurrentStateOfResource(apiResource, gvr, actualTyped.Name, actualTyped.Namespace)
	return actInstance, err
}

func getFieldByPath(obj interface{}, path []string) (interface{}, bool) {
	unstructObj := obj.(*unstructured.Unstructured)

	var mapObj interface{}
	mapObj = unstructObj.Object

	lastProcessedPathIdx := -1
	for i, element := range path {
		num, err := strconv.Atoi(element)
		if err == nil {
			slice, found, _ := unstructured.NestedSlice(mapObj.(map[string]interface{}), extractPath(path, lastProcessedPathIdx+1, i-1)...)
			if !found {
				return nil, false
			}
			mapObj = slice[num]
			lastProcessedPathIdx = i
		}
	}

	var field interface{}
	var found bool

	switch mapObj.(type) {
	case map[string]interface{}:
		field, found, _ = unstructured.NestedFieldCopy(mapObj.(map[string]interface{}), extractPath(path, lastProcessedPathIdx+1, len(path)-1)...)
	default:
		found = true
		field = mapObj
	}

	return field, found
}

func extractPath(path []string, fromIdx int, toIdx int) []string {
	if toIdx > len(path)-1 || fromIdx < 0 || fromIdx > toIdx {
		return nil
	}
	var retArr []string

	for i := fromIdx; i <= toIdx; i++ {
		retArr = append(retArr, path[i])
	}

	return retArr
}
