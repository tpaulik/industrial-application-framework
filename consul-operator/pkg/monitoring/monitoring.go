// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package monitoring

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"

	app "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	kubelib2 "github.com/nokia/industrial-application-framework/consul-operator/libs/kubelib"

	"github.com/nokia/industrial-application-framework/alarmlogger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Monitor struct {
	RuntimeClient      client.Client
	Instance           *app.Consul
	Namespace          string
	ClientSet          *kubernetes.Clientset
	Running            bool
	RunningCallback    func()
	NotRunningCallback func()
	pauseChannel       chan struct{}
}

var (
	log                        = logf.Log.WithName("monitoring_controller")
	monitoringInstance         *Monitor
	isAppNotRunningAlarmActive bool
)

func NewMonitor(runtimeClient client.Client, instance *app.Consul, namespace string,
	runningCallback func(), notRunningCallback func()) *Monitor {
	if monitoringInstance == nil {
		monitoringInstance = &Monitor{
			RuntimeClient:      runtimeClient,
			Instance:           instance,
			Namespace:          namespace,
			ClientSet:          kubelib2.GetKubeAPI(),
			RunningCallback:    runningCallback,
			NotRunningCallback: notRunningCallback,
		}
	}
	return monitoringInstance
}

func (m *Monitor) Run() {
	if m.Running {
		return
	}
	m.Running = true

	log.Info("Watching application")

	m.pauseChannel = make(chan struct{})

	go m.watchInformer(
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				log.Info("Pod deleted")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Info("Pod changed")

				status := m.GetApplicationStatus()
				if m.Instance.Status.AppStatus != status {
					switch status {
					case app.AppStatusRunning:
						if isAppNotRunningAlarmActive {
							// clear alarm
							alarmlogger.ClearAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
								Name:     "AppNotRunning",
								ID:       "1",
								Severity: alarmlogger.Warning,
								Text:     "All components are now ready",
							})
							isAppNotRunningAlarmActive = false
						}
						m.RunningCallback()
					case app.AppStatusNotRunning:
						if !isAppNotRunningAlarmActive {
							// raise alarm
							alarmlogger.RaiseAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
								Name:     "AppNotRunning",
								ID:       "1",
								Severity: alarmlogger.Warning,
								Text:     "Not all components are ready",
							})
							isAppNotRunningAlarmActive = true
						}
						m.NotRunningCallback()
					}
				}

				m.Instance.Status.AppStatus = status
				if err := m.updateAppStatus(m.Instance); nil != err {
					log.Error(err, "status appStatus update failed")
				}

				log.Info("UpdateFunc", "status", m.Instance.Status.AppStatus)
			},
			AddFunc: func(obj interface{}) {},
		}, m.pauseChannel)
}

func (m *Monitor) updateAppStatus(instance *app.Consul) error {
	appStatus := instance.Status.AppStatus
	key := client.ObjectKey{
		Namespace: instance.GetNamespace(),
		Name:      instance.GetName(),
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := m.RuntimeClient.Get(context.TODO(), key, instance)
		if err != nil {
			return err
		}
		instance.Status.AppStatus = appStatus
		err = m.RuntimeClient.Status().Update(context.TODO(), instance)
		return err
	})

	if err != nil {
		return errors.Wrap(err, "failed app status update")
	}

	return nil
}


func (m *Monitor) Pause() {
	if m.Running {
		log.Info("Watching application paused")
		m.Running = false
		close(m.pauseChannel)
	}
}

func (m *Monitor) GetApplicationStatus() app.AppStatus {
	pods, _ := m.ClientSet.CoreV1().Pods(m.Namespace).List(context.TODO(), v1.ListOptions{LabelSelector: "statusCheck=true"})
	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) == 0 {
			return app.AppStatusNotRunning
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				return app.AppStatusNotRunning
			}
		}
		return app.AppStatusRunning
	}
	return app.AppStatusNotRunning
}

func (m *Monitor) watchInformer(eventHandler cache.ResourceEventHandler, stopper chan struct{}) {
	listOptionsFunc := internalinterfaces.TweakListOptionsFunc(func(options *v1.ListOptions) {
		options.LabelSelector = "statusCheck=true"
	})

	informer := informersv1.NewFilteredPodInformer(
		m.ClientSet,
		m.Namespace,
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		listOptionsFunc,
	)

	informer.AddEventHandler(eventHandler)
	informer.Run(stopper)

	log.Info("Application watch has been stopped")
}
