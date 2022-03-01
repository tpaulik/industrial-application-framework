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
	appsv1 "k8s.io/api/apps/v1"
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

var k8sClient client.Client
var consulOperatorCr *appdacnokiacomv1alpha1.Consul

type resource struct {
	resourceName   string
	kind           string
	statusContents map[string]interface{}
}

var _ = Describe("deploy/undeploy case", func() {
	BeforeEach(func() {
	}, 60)

	AfterEach(func() {
	}, 60)

	var err error

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

	var consulCrInstance *appdacnokiacomv1alpha1.Consul

	Describe("deploy case", func() {
		Context("the appframework", func() { //actor
			It("creates a test namespace", func() { //mit csinál
				createNameSpace(consulTestNamespace)
			})
			It("executes consul CR", func() {
				consulCrInstance = getConsulCrInstance()
				os.Setenv("DEPLOYMENT_DIR", "../deployment")
				os.Setenv("RESREQ_DIR", "../deployment/resource-reqs-generated")
				os.Setenv("KUBECONFIG", ctenv.LocalCfg.KubeConfig)
				log.Info("Suite Stories", "Kubeconfig file", ctenv.LocalCfg.KubeConfig)

				deploymentAbsPath := getTestBinaryPath("/../deployment")
				path := os.Getenv("PATH")
				os.Setenv("PATH", path+":"+deploymentAbsPath)
				err = k8sClient.Create(context.TODO(), consulCrInstance)

				Expect(err).NotTo(HaveOccurred())
			})
			It("executes the resource CRs and they get approved", func() {

				Expect(createResourceCr(metricsEndpoint.resourceName, consulTestNamespace, metricsEndpoint.kind)).To(ExistsK8sRes(5 * time.Second))
				Expect(createResourceCr(privateNetwork.resourceName, consulTestNamespace, privateNetwork.kind)).To(ExistsK8sRes(5 * time.Second))
				Expect(createResourceCr(resourceRequest.resourceName, consulTestNamespace, resourceRequest.kind)).To(ExistsK8sRes(5 * time.Second))
				Expect(createResourceCr(storage.resourceName, consulTestNamespace, storage.kind)).To(ExistsK8sRes(5 * time.Second))

				approveResource(metricsEndpoint)
				approveResource(privateNetwork)
				approveResource(resourceRequest)
				approveResource(storage)

			})
			It("checks if the stateful set is present", func() {

				Eventually(func() error {
					_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Get(context.TODO(), "example-consul", metav1.GetOptions{})
					return err
				}, 35*time.Second, time.Second*1).Should(BeNil())
			})
			It("updates the stateful set to contain initContainers", func() {
				var consulStatefulSet *appsv1.StatefulSet
				consulStatefulSet, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Get(context.TODO(), "example-consul", metav1.GetOptions{})

				consulStatefulSet.Spec.Template.Spec.InitContainers = []corev1.Container{corev1.Container{
					Name:  "appfw-private-network-routing",
					Image: "registry.dac.nokia.com/public/calico/node:v3.18.2",
					Args:  []string{initcontainerArgs},
				}}

				kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Update(context.TODO(), consulStatefulSet, metav1.UpdateOptions{})
			})
			It("creates a dummy pod with consul name and sets it to running", func() {
				fakePodCr := getStaticPodCr()
				err = k8sClient.Create(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{corev1.ContainerStatus{Ready: true}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appStatus"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Expect(consulCrResourceId).To(EqualsK8sRes("RUNNING", 10*time.Second))
			})
			It("updates stateful set and checks if the app reported data shows the fixed pod Ip", func() {
				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appReportedData", "privateNetworkIpAddresses", "statefulsets/example-consul"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Eventually(consulCrResourceId, 10*time.Second, time.Second*1).Should(EqualsK8sRes("10.32.1.2"))
			})
		})
		Describe("License Expired Case", func() {
			It("creates the LicenceExpiration resource in the namespace of the application", func() {

				var licenseExpiredCr = unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "ops.dac.nokia.com/v1alpha1",
						"kind":       "LicenceExpired",
						"metadata":   map[string]interface{}{"name": "consul-op-license-expired"},
					},
				}

				licenseExpiredGvr := schema.GroupVersionResource{Group: gvkResourceGroup, Version: gvkVersion, Resource: "licenceexpireds"}

				_, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(licenseExpiredGvr).Namespace(consulTestNamespace).Create(context.TODO(), &licenseExpiredCr, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
			It("stops the consul service", func() {
				//_, err = kubelib.GetKubeAPI().CoreV1().Services(consulTestNamespace).Get(context.TODO(), "example-consul-service", metav1.GetOptions{})

				//Eventually(k8serrors.IsNotFound(err), 10*time.Second, time.Second*1).Should(BeTrue())
			})
			It("changes the appStatus to Frozen in the app spec CRs", func() {
				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appStatus"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Expect(consulCrResourceId).To(EqualsK8sRes("FROZEN", 10*time.Second))
			})
			It("but keeps the stateful set intact", func() {
				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Get(context.TODO(), "example-consul", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Describe("License expiration lifted case", func() {
			It("The LicenceExpiration resource is removed from the namespace of the application", func() {
				licenseExpiredGvr := schema.GroupVersionResource{Group: gvkResourceGroup, Version: gvkVersion, Resource: "licenceexpireds"}

				err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(licenseExpiredGvr).Namespace(consulTestNamespace).Delete(context.TODO(), "consul-op-license-expired", metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
			It("the appStatus is changed back to running", func() {
				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appStatus"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Expect(consulCrResourceId).To(EqualsK8sRes("RUNNING", 35*time.Second))
			})
		})
		Describe("app status monitoring case", func() {
			It("The pod status is set to failed", func() {
				fakePodCr := getStaticPodCr()

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{corev1.ContainerStatus{Ready: false}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())
			})
			It("the consul application detects the stopped pod and removes its running state", func() {
				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appStatus"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Expect(consulCrResourceId).To(EqualsK8sRes("NOT_RUNNING", 10*time.Second))
			})
			It("Once the pot status is set back to running, the application resolved its running state as well", func() {
				fakePodCr := getStaticPodCr()

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{corev1.ContainerStatus{Ready: true}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: consulTestNamespace,
					ParamPath: []string{"status", "appStatus"},
					Gvk: schema.GroupVersionKind{
						Group:   gvkAppGroup,
						Version: gvkVersion,
						Kind:    gvkConsulKind,
					},
				}
				Expect(consulCrResourceId).To(EqualsK8sRes("RUNNING", 10*time.Second))
			})

		})
		Describe("App reported data case", func() {

		})
		Describe("undeploy case", func() {
			It("undeploys the operator", func() {
				err = k8sClient.Delete(context.TODO(), consulCrInstance)
				Expect(err).NotTo(HaveOccurred())

				fakePodCr := getStaticPodCr()
				err = k8sClient.Delete(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

			})
			It("removes the finalizer from the app CR and let kubernetes delete the app CR", func() {
				Eventually(func() error {
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: consulTestNamespace, Name: consulAppName}, consulCrInstance)
					return err
				}, 10*time.Second, time.Second*1).Should(HaveOccurred())
			})
			It("checks if the Stateful set and all resources are removed", func() {
				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(consulTestNamespace).Get(context.TODO(), "example-consul", metav1.GetOptions{})

				Eventually(k8serrors.IsNotFound(err), 10*time.Second, time.Second*1).Should(BeTrue())

				checkIfResourceDoesNotExist(metricsEndpoint.resourceName, metricsEndpoint.kind)
				checkIfResourceDoesNotExist(privateNetwork.resourceName, privateNetwork.kind)
				checkIfResourceDoesNotExist(resourceRequest.resourceName, resourceRequest.kind)
				checkIfResourceDoesNotExist(storage.resourceName, storage.kind)

			})

		})
	})

})

func createNameSpace(namespace string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	err := k8sClient.Create(context.TODO(), ns)
	Expect(err).NotTo(HaveOccurred())
}

func approveResource(resourceDef resource) {
	var gvr schema.GroupVersionResource
	var resourceSpecCr *unstructured.Unstructured
	var err error

	gvr, _, err = GetGvrAndAPIResources(schema.GroupVersionKind{Group: gvkResourceGroup, Version: gvkVersion, Kind: resourceDef.kind})
	Expect(err).NotTo(HaveOccurred())

	resourceSpecCr, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(consulTestNamespace).Get(context.TODO(), resourceDef.resourceName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	resourceSpecCr.Object["status"] = resourceDef.statusContents

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

func checkIfResourceDoesNotExist(resourceCrName string, kind string) {
	var gvr schema.GroupVersionResource
	var err error

	gvr, _, err = GetGvrAndAPIResources(schema.GroupVersionKind{Group: gvkResourceGroup, Version: gvkVersion, Kind: kind})
	Expect(err).NotTo(HaveOccurred())

	_, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(consulTestNamespace).Get(context.TODO(), resourceCrName, metav1.GetOptions{})
	Expect(k8serrors.IsNotFound(err)).To(BeTrue())
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
