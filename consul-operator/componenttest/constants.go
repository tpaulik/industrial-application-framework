package componenttest

import (
	appdacnokiacomv1alpha1 "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

const (
	apiVersion    = "app.dac.nokia.com/v1alpha1"
	consulAppName = "consul-app"

	approved = "Approved"

	altPort   = 8400
	consulDns = 8600
	httpPort  = 8080
	httpsPort = 8443
	serflan   = 8301
	serfwan   = 8302
	server    = 8300
	udpPort   = 53
	uiPort    = 8500

	replicaCount         = 1
	metricsDomain        = "metrics.consul.appdomain.com"
	networkInterfaceName = "consul-if"

	apnUUID     = "apn:anyApnUUID"
	appPodFixIp = "127.0.2.0/32"

	operatorDefaultWaitTimeout = 5 * time.Second
	consulTestNamespace        = "consul-test-ns"

	gvkResourceGroup = "ops.dac.nokia.com"

	gvkAppGroup   = "app.dac.nokia.com"
	gvkVersion    = "v1alpha1"
	gvkConsulKind = "Consul"
)

var consulGvk = schema.GroupVersionKind{Group: "app.dac.nokia.com", Version: "v1alpha1", Kind: "Consul"}

var ports = appdacnokiacomv1alpha1.Ports{
	UiPort:    uiPort,
	AltPort:   altPort,
	UdpPort:   udpPort,
	HttpPort:  httpPort,
	HttpsPort: httpsPort,
	Serflan:   serflan,
	Serfwan:   serfwan,
	ConsulDns: consulDns,
	Server:    server,
}

var networks = []appdacnokiacomv1alpha1.Network{{ApnUUID: apnUUID, AdditionalRoutes: []string{"127.0.1.1/28", "127.0.1.2/28"}}}

func getConsulCrInstance() *appdacnokiacomv1alpha1.Consul {
	consulCrInstance := appdacnokiacomv1alpha1.Consul{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvkConsulKind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      consulAppName,
			Namespace: consulTestNamespace,
		},
		Spec: appdacnokiacomv1alpha1.ConsulSpec{
			ReplicaCount:      replicaCount,
			Ports:             ports,
			MetricsDomainName: metricsDomain,
			PrivateNetworkAccess: &appdacnokiacomv1alpha1.PrivateNetworkAccess{
				Networks:             networks,
				CustomerNetwork:      "127.0.0.0/28",
				NetworkInterfaceName: networkInterfaceName,
				AppPodFixIp: &appdacnokiacomv1alpha1.AppPodFixIp{
					Db: appPodFixIp,
				},
			},
		},
	}

	return &consulCrInstance
}

func getStaticPodCr() *corev1.Pod {
	staticPodCr := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-consul",
			Namespace: consulTestNamespace,
			Labels:    map[string]string{"app": "example-consul", "statusCheck": "true"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{corev1.Container{
				Name:  "example-consul",
				Image: "registry.dac.nokia.com/public/consul:1.4.4",
				Ports: []corev1.ContainerPort{corev1.ContainerPort{
					ContainerPort: 80,
					Protocol:      "TCP",
				}},
			}},
		},
	}
	return staticPodCr
}
