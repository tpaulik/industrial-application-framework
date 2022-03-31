// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package platformres

import (
	"fmt"
	k8sdynamic2 "github.com/nokia/industrial-application-framework/application-lib/pkg/k8sdynamic"
	kubelib2 "github.com/nokia/industrial-application-framework/application-lib/pkg/kubelib"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"strings"
	"sync"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("platformres")

const (
	StatusField         = "status"
	ApprovalStatusField = "approvalStatus"
)

func ApplyPlatformResourceRequests(namespace string, resourceRequestPath string) ([]k8sdynamic2.ResourceDescriptor, error) {
	logger := log.WithName("ApplyPlatformResourceRequests")
	logger.Info("Called")

	dynClient := k8sdynamic2.New(kubelib2.GetKubeAPI())
	dir := resourceRequestPath
	if dir == "" {
		return nil, errors.New("resReqDir is not set")
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read dir: %v", dir)
	}

	var descList []k8sdynamic2.ResourceDescriptor
	for _, file := range files {
		if !file.IsDir() {
			fileContent, err := ioutil.ReadFile(dir + "/" + file.Name())
			if err != nil {
				return nil, errors.Wrap(err, "failed to read file")
			}

			if strings.TrimSpace(string(fileContent)) == "" {
				logger.Info("File is empty skip it", "path", dir+"/"+file.Name())
				continue
			}
			resourceDesc, err := dynClient.ApplyYamlResource(string(fileContent), namespace)
			descList = append(descList, resourceDesc)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply the request in k8s")
			}
		}
	}

	return descList, nil
}
func ApplyPnaResourceRequests(namespace string, resourceRequestPath string) ([]k8sdynamic2.ResourceDescriptor, error) {
	logger := log.WithName("ApplyPnaResourceRequests")
	logger.Info("Called")

	dynClient := k8sdynamic2.New(kubelib2.GetKubeAPI())
	dir := resourceRequestPath
	if dir == "" {
		return nil, errors.New("resReqDir is not set")
	}

	var descList []k8sdynamic2.ResourceDescriptor
	fileContent, err := ioutil.ReadFile(dir + "/private-network-access.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	resourceDesc, err := dynClient.ApplyYamlResource(string(fileContent), namespace)
	descList = append(descList, resourceDesc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply the request in k8s")
	}
	return descList, nil
}

func WaitUntilResourcesGranted(resourceList []k8sdynamic2.ResourceDescriptor, timeout time.Duration) error {
	logger := log.WithName("WaitUntilResourcesGranted")

	var stopperList = sync.Map{}
	var waitGroup sync.WaitGroup
	var results []*bool

	for _, resource := range resourceList {
		stopper := make(chan struct{})
		waitGroup.Add(1)
		var result bool
		results = append(results, &result)
		stopperList.Store(resource.Name, stopper)
		startWatchResourceRequest(
			resource.Name,
			resource.Namespace,
			"",
			resource.Gvr.GetGvr(),
			stopper,
			&stopperList,
			&waitGroup,
			&result,
		)
	}

	if waitTimeout(&waitGroup, timeout) {
		var timedOutResources []string
		stopperList.Range(func(resName, stopper interface{}) bool {
			timedOutResources = append(timedOutResources, resName.(string))
			close(stopper.(chan struct{}))
			return true
		})
		return errors.New(fmt.Sprintf("waiting for the approval of the following platform resource requests timed out: %v", timedOutResources))
	} else {
		for _, result := range results {
			if !*result {
				return errors.New("some of the platform resource request(s) has rejected")
			}
		}
	}
	logger.Info("All of the requested platform resources have been granted")
	return nil
}

const (
	ApprovalStatusApproved = "Approved"
	ApprovalStatusRejected = "Rejected"
)

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		wg.Wait()
	}()
	select {
	case <-finished:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func startWatchResourceRequest(name string, namespace string, resourceVersion string, gvr schema.GroupVersionResource, stopper chan struct{}, stopperList *sync.Map, waitGroup *sync.WaitGroup, result *bool) {
	logger := log.WithName("StartWatchResourceRequest").WithValues("resource", name)

	logger.Info("Watching resource")

	go k8sdynamic2.WatchInformer(
		name, namespace, resourceVersion, gvr,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				logger.V(1).Info("Delete resource detected")
				close(stopper)
				waitGroup.Done()
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				logger.V(1).Info("Resource request modification detected")

				newValue, _ := getApprovalStatus(newObj)
				oldValue, _ := getApprovalStatus(oldObj)

				if oldValue != newValue {
					switch newValue {
					case ApprovalStatusApproved:
						*result = true
						logger.Info("Resource approved")
						break
					default:
						*result = false
						logger.Info("Cannot create resource")
						break
					}
					waitGroup.Done()
					close(stopper)
					stopperList.Delete(name)
				}
			},
			AddFunc: func(obj interface{}) {
				logger.V(1).Info("Add resource detected")
				value, _ := getApprovalStatus(obj)
				if value == ApprovalStatusApproved {
					*result = true
					waitGroup.Done()
					close(stopper)
					stopperList.Delete(name)
				}
			},
		},
		stopper)
}
func getApprovalStatus(obj interface{}) (string, bool) {
	unstructObj := obj.(*unstructured.Unstructured)
	value, found, _ := unstructured.NestedString(unstructObj.Object, StatusField, ApprovalStatusField)
	return value, found
}
