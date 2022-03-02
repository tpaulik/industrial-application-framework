package componenttest

import (
	. "github.com/nokia/industrial-application-framework/componenttest-lib/pkg/matcher"
	appdacnokiacomv1alpha1 "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

const (
	apiVersion            = "app.dac.nokia.com/v1alpha1"
	consulAppName         = "consul-app"
	consulStatefulSetName = "example-consul"

	appPodFixIp = "10.32.1.2"

	initcontainerArgs = "iptables -t nat -A POSTROUTING -o appfw-appnet13 -j SNAT --to-source " + appPodFixIp + " " +
		"&& ip a && mkdir -p /etc/iproute2 && touch /etc/iproute2/rt_tables && echo 200 custom >> /etc/iproute2/rt_tables && ip rule add from " + appPodFixIp + "/32 lookup custom && ip route add default via 169.254.151.193 dev appfw-appnet13 table custom " +
		"&& ip route add 192.168.245.0/24 via 169.254.151.193 dev appfw-appnet13 && ip link add name private-net type dummy && ip addr add " + appPodFixIp + "/32 brd + dev private-net && ip link set private-net up"

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

	apnUUID = "apn:anyApnUUID"

	defaultWaitTimeout = 5 * time.Second
	testNamespace      = "consul-test-ns"

	opsGroup  = "ops.dac.nokia.com"
	appsGroup = "app.dac.nokia.com"

	groupResourceVersion = "v1alpha1"
	consulKind           = "Consul"
)

var expectedPorts = []int32{altPort, consulDns, httpPort, httpsPort, serflan, serfwan, server, uiPort} // udp/53 is not included

var consulGvk = schema.GroupVersionKind{Group: appsGroup, Version: groupResourceVersion, Kind: consulKind}

var consulAppStatusResourceId = K8sResourceId{
	Name:      consulAppName,
	Namespace: testNamespace,
	ParamPath: []string{"status", "appStatus"},
	Gvk:       consulGvk,
}

type resource struct {
	resourceName   string
	kind           string
	statusContents map[string]interface{}
}

var metricsEndpoint = resource{
	resourceName: "consul-metricsendpoint",
	kind:         "MetricsEndpoint",
	statusContents: map[string]interface{}{
		"approvalStatus": approved,
	},
}

var privateNetwork = resource{
	resourceName: "private-network-for-consul",
	kind:         "PrivateNetworkAccess",
	statusContents: map[string]interface{}{
		"approvalStatus": approved,
		"assignedNetwork": map[string]interface{}{
			"name": "anyDummyNetwork",
		},
	},
}

var resourceRequest = resource{
	resourceName: "resource-for-consul",
	kind:         "Resourcerequest",
	statusContents: map[string]interface{}{
		"approvalStatus": approved,
	},
}

var storage = resource{
	resourceName: "storage-for-db",
	kind:         "Storage",
	statusContents: map[string]interface{}{
		"approvalStatus": approved,
	},
}

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
			Kind:       consulKind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      consulAppName,
			Namespace: testNamespace,
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

var initContainers = []corev1.Container{corev1.Container{
	Name:  "appfw-private-network-routing",
	Image: "registry.dac.nokia.com/public/calico/node:v3.18.2",
	Args:  []string{initcontainerArgs},
}}

func getStaticPodCr() *corev1.Pod {
	staticPodCr := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-consul",
			Namespace: testNamespace,
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
			InitContainers: initContainers,
		},
	}
	return staticPodCr
}
