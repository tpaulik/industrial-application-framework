package componenttest

import (
	"context"
	ctenv "github.com/nokia/industrial-application-framework/componenttest-lib/pkg/env"
	ctk8sclient "github.com/nokia/industrial-application-framework/componenttest-lib/pkg/k8sclient"
	. "github.com/nokia/industrial-application-framework/componenttest-lib/pkg/matcher"
	appdacnokiacomv1alpha1 "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/libs/kubelib"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var log = logf.Log.WithName("consulTests")

const (
	apiVersion   = "app.dac.nokia.com/v1alpha1"
	metadataName = "consul-app"

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

var k8sClient client.Client
var consulOperatorCr *appdacnokiacomv1alpha1.Consul

var _ = Describe("consul-operator", func() {
	BeforeEach(func() {
	}, 60)

	AfterEach(func() {
	}, 60)

	var err error
	var resourceKindByCr = map[string]string{
		"consul-metricsendpoint":     "MetricsEndpoint",
		"private-network-for-consul": "PrivateNetworkAccess",
		"resource-for-consul":        "Resourcerequest",
		"storage-for-db":             "Storage",
	}
	var consulCrInstance *appdacnokiacomv1alpha1.Consul

	Context("the consul operator", func() {
		It("creates consul test namespace", func() {
			createNameSpace(consulTestNamespace)
		})
		It("sets up consul CR and executes it", func() {
			consulCrInstance = getConsulCrInstance()
			os.Setenv("DEPLOYMENT_DIR", "../deployment")
			os.Setenv("RESREQ_DIR", "../deployment/resource-reqs-generated")
			os.Setenv("KUBECONFIG", ctenv.LocalCfg.KubeConfig)
			log.Info("Suite Stories", "Kubeconfig file", ctenv.LocalCfg.KubeConfig)

			deploymentAbsPath := getTestBinaryPath("/../deployment")
			path := os.Getenv("PATH")
			os.Setenv("PATH", path+":"+deploymentAbsPath)
			err = k8sClient.Create(context.TODO(), consulCrInstance)

			for cr, kind := range resourceKindByCr {
				Expect(createResourceCr(cr, consulTestNamespace, kind)).To(ExistsK8sRes(10 * time.Second))
			}

			Expect(err).NotTo(HaveOccurred())
		})
		It("approves the resource requests", func() {
			for cr, kind := range resourceKindByCr {
				approveResource(cr, kind)
			}
		})
		It("checks if the stateful set is present", func() {

			Eventually(func() error {
				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Get(context.TODO(), "example-consul", metav1.GetOptions{})
				return err
			}, 35*time.Second, time.Second*1).Should(BeNil())
		})
	})

})

func getConsulCrInstance() *appdacnokiacomv1alpha1.Consul {
	consulCrInstance := appdacnokiacomv1alpha1.Consul{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvkConsulKind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadataName,
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

func createNameSpace(namespace string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	err := k8sClient.Create(context.TODO(), ns)
	Expect(err).NotTo(HaveOccurred())
}

func approveResource(resourceCrName string, kind string) {
	var gvr schema.GroupVersionResource
	var resourceSpecCr *unstructured.Unstructured
	var err error

	gvr, _, err = GetGvrAndAPIResources(schema.GroupVersionKind{Group: gvkResourceGroup, Version: gvkVersion, Kind: kind})
	Expect(err).NotTo(HaveOccurred())

	resourceSpecCr, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(consulTestNamespace).Get(context.TODO(), resourceCrName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	resourceSpecCr.Object["status"] = map[string]interface{}{
		"approvalStatus": approved,
	}

	resourceSpecCr, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(consulTestNamespace).UpdateStatus(context.TODO(), resourceSpecCr, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())

}

func createResourceCr(name string, namespace string, kind string) K8sResourceId {
	return K8sResourceId{
		Name:      name,
		Namespace: namespace,
		Gvk:       schema.GroupVersionKind{Group: gvkResourceGroup, Version: gvkVersion, Kind: kind},
	}
}

func deleteNameSpace(namespace string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), ns)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: namespace}, ns)
		if k8serrors.IsNotFound(err) {
			log.Info("Namespace removed")
			return err
		}
		return nil
	}, operatorDefaultWaitTimeout).Should(HaveOccurred())

}
