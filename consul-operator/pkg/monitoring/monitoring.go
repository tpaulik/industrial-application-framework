package monitoring

import (
	"context"

	kubelib2 "github.com/nokia/industrial-application-framework/consul-operator/libs/kubelib"
	dac "github.com/nokia/industrial-application-framework/consul-operator/pkg/apis/dac/v1alpha2"

	"gitlabe2.ext.net.nokia.com/ndac-appfw/alarmlogger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type Monitor struct {
	RuntimeClient      client.Client
	Instance           *dac.Consul
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

func NewMonitor(runtimeClient client.Client, instance *dac.Consul, namespace string,
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
					case dac.AppStatusRunning:
						if isAppNotRunningAlarmActive {
							// clear alarm
							alarmlogger.ClearAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
								Name:     "AppNotRunning",
								ID: 	  "1",
								Severity: alarmlogger.Warning,
								Text:     "All components are now ready",
							})
							isAppNotRunningAlarmActive = false
						}
						m.RunningCallback()
					case dac.AppStatusNotRunning:
						if !isAppNotRunningAlarmActive {
							// raise alarm
							alarmlogger.RaiseAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{
								Name:     "AppNotRunning",
								ID: 	  "1",
								Severity: alarmlogger.Warning,
								Text:     "Not all components are ready",
							})
							isAppNotRunningAlarmActive = true
						}
						m.NotRunningCallback()
					}
				}

				m.Instance.Status.AppStatus = status
				if err := m.RuntimeClient.Status().Update(context.TODO(), m.Instance); nil != err {
					log.Error(err, "status appStatus update failed")
				}

				log.Info("UpdateFunc", "status", m.Instance.Status.AppStatus)
			},
			AddFunc: func(obj interface{}) {},
		}, m.pauseChannel)
}

func (m *Monitor) Pause() {
	if m.Running {
		log.Info("Watching application paused")
		m.Running = false
		close(m.pauseChannel)
	}
}

func (m *Monitor) GetApplicationStatus() dac.AppStatus {
	pods, _ := m.ClientSet.CoreV1().Pods(m.Namespace).List(v1.ListOptions{LabelSelector: "statusCheck=true"})
	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) == 0 {
			return dac.AppStatusNotRunning
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				return dac.AppStatusNotRunning
			}
		}
		return dac.AppStatusRunning
	}
	return dac.AppStatusNotRunning
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
