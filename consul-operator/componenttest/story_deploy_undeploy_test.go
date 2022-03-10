// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

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

var _ = Describe("Consul Operator Component Tests", func() {
	BeforeEach(func() {
	}, 60)

	AfterEach(func() {
	}, 60)

	var err error

	var consulCrInstance *appdacnokiacomv1alpha1.Consul
	var consulStatefulSet *appsv1.StatefulSet

	Describe("Consul Operator deploy case", func() {
		Context("The Application Framework", func() {
			It("creates a test namespace", func() {
				createNameSpace(testNamespace)
			})
		})
		Context("The Test Environment", func() {
			It("sets the necessary environment variables", func() {

				os.Setenv("DEPLOYMENT_DIR", "../deployment")
				os.Setenv("RESREQ_DIR", "../deployment/resource-reqs-generated")
				os.Setenv("KUBECONFIG", ctenv.LocalCfg.KubeConfig)

				log.Info("Kubernetes Temporary Location", "Kubeconfig file", ctenv.LocalCfg.KubeConfig)

				deploymentAbsPath := getTestBinaryPath("/../deployment")
				path := os.Getenv("PATH")
				os.Setenv("PATH", path+":"+deploymentAbsPath)
			})
		})
		Context("The Application Framework", func() {
			It("executes consul CR", func() {

				consulCrInstance = getConsulCrInstance()
				err = k8sClient.Create(context.TODO(), consulCrInstance)

				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("The Consul Operator", func() {
			It("executes the resource CRs", func() {

				Expect(createResourceCr(metricsEndpoint.resourceName, testNamespace, metricsEndpoint.kind)).To(ExistsK8sRes(consulTestDefaultTimeout))
				Expect(createResourceCr(privateNetwork.resourceName, testNamespace, privateNetwork.kind)).To(ExistsK8sRes(consulTestDefaultTimeout))
				Expect(createResourceCr(resourceRequest.resourceName, testNamespace, resourceRequest.kind)).To(ExistsK8sRes(consulTestDefaultTimeout))
				Expect(createResourceCr(storage.resourceName, testNamespace, storage.kind)).To(ExistsK8sRes(consulTestDefaultTimeout))
			})
		})
		Context("The Application Framework", func() {
			It("approves the resources", func() {
				approveResource(metricsEndpoint)
				approveResource(privateNetwork)
				approveResource(resourceRequest)
				approveResource(storage)
			})
			It("checks if the stateful set is present", func() {
				Eventually(func() error {
					_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(testNamespace).Get(context.TODO(), consulStatefulSetName, metav1.GetOptions{})
					return err
				}, 35*time.Second, time.Second*1).Should(BeNil())
			})
			It("updates the stateful set to contain initContainer values", func() {

				consulStatefulSet, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(testNamespace).Get(context.TODO(), consulStatefulSetName, metav1.GetOptions{})

				consulStatefulSet.Spec.Template.Spec.InitContainers = initContainers

				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(testNamespace).Update(context.TODO(), consulStatefulSet, metav1.UpdateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
			It("creates a dummy pod for the Consul Application and sets it to running", func() {
				fakePodCr := getStaticPodCr()
				err = k8sClient.Create(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

				Expect(consulAppStatusResourceId).To(EqualsK8sRes("RUNNING", 10*time.Second))
			})
			It("checks if the app reported data shows the fixed pod Ip", func() {
				consulCrResourceId := K8sResourceId{
					Name:      consulAppName,
					Namespace: testNamespace,
					ParamPath: []string{"status", "appReportedData", "privateNetworkIpAddresses", "statefulsets/" + consulStatefulSetName},
					Gvk:       consulGvk,
				}
				Eventually(consulCrResourceId, consulTestDefaultTimeout, time.Second*1).Should(EqualsK8sRes(appPodFixIp))
			})
			It("checks if the ports given in the CR are present in the service", func() {
				var service *corev1.Service
				service, err = kubelib.GetKubeAPI().CoreV1().Services(testNamespace).Get(context.TODO(), consulServiceName, metav1.GetOptions{})

				missingPorts := findMissingPorts(service.Spec.Ports)

				Expect(missingPorts).To(BeEmpty())
			})
		})
	})

	Describe("License Expired Case", func() {
		Context("The Application Framework", func() {
			It("creates the LicenceExpiration resource in the namespace of the application", func() {

				var licenseExpiredCr = unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "ops.dac.nokia.com/v1alpha1",
						"kind":       "LicenceExpired",
						"metadata":   map[string]interface{}{"name": "consul-op-license-expired"},
					},
				}

				licenseExpiredGvr := schema.GroupVersionResource{Group: opsGroup, Version: groupResourceVersion, Resource: "licenceexpireds"}

				_, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(licenseExpiredGvr).Namespace(testNamespace).Create(context.TODO(), &licenseExpiredCr, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("The Consul Operator", func() {
			It("changes the appStatus to Frozen", func() {

				Expect(consulAppStatusResourceId).To(EqualsK8sRes("FROZEN", consulTestDefaultTimeout))
			})
			It("But keeps the stateful set intact", func() {
				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(testNamespace).Get(context.TODO(), consulStatefulSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("License activated case", func() {
		Context("The Application Framework", func() {
			It("Removes the LicenceExpiration resource from the namespace of the application", func() {
				licenseExpiredGvr := schema.GroupVersionResource{Group: opsGroup, Version: groupResourceVersion, Resource: "licenceexpireds"}

				err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(licenseExpiredGvr).Namespace(testNamespace).Delete(context.TODO(), "consul-op-license-expired", metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("The Consul Operator", func() {
			It("Sets the appStatus is to running", func() {
				Expect(consulAppStatusResourceId).To(EqualsK8sRes("RUNNING", consulTestDefaultTimeout))
			})
		})
	})

	Describe("Application Pod failure case", func() {
		Context("The Application framework", func() {
			It("sets the pod status to failed", func() {
				fakePodCr := getStaticPodCr()

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{{Ready: false}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("The Consul Application", func() {
			It("detects the stopped pod and removes its running state", func() {
				Expect(consulAppStatusResourceId).To(EqualsK8sRes("NOT_RUNNING", consulTestDefaultTimeout))
			})
		})
	})

	Describe("The Application pod resumes operation", func() {
		Context("The Application Framework", func() {
			It("Resumes pod status back to running", func() {
				fakePodCr := getStaticPodCr()

				fakePodCr.Status = corev1.PodStatus{
					Phase:             corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				}

				err = k8sClient.Status().Update(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("The Consul operator", func() {
			It("Resumes it status to running", func() {
				Expect(consulAppStatusResourceId).To(EqualsK8sRes("RUNNING", consulTestDefaultTimeout))
			})
		})
	})

	Describe("Consul Operator undeploy case", func() {
		Context("The Application Framework", func() {
			It("Deletes the operator CR and application pod", func() {
				err = k8sClient.Delete(context.TODO(), consulCrInstance)
				Expect(err).NotTo(HaveOccurred())

				fakePodCr := getStaticPodCr()
				err = k8sClient.Delete(context.TODO(), fakePodCr)
				Expect(err).NotTo(HaveOccurred())

			})
			It("removes the finalizer from the app CR and let kubernetes delete the app CR", func() {
				Eventually(func() error {
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: testNamespace, Name: consulAppName}, consulCrInstance)
					return err
				}, consulTestDefaultTimeout, time.Second*1).Should(HaveOccurred())
			})
			It("checks if the Stateful set and all resources are removed", func() {
				_, err = kubelib.GetKubeAPI().AppsV1().StatefulSets(testNamespace).Get(context.TODO(), consulStatefulSetName, metav1.GetOptions{})

				Eventually(k8serrors.IsNotFound(err), consulTestDefaultTimeout, time.Second*1).Should(BeTrue())

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

	gvr, _, err = GetGvrAndAPIResources(schema.GroupVersionKind{Group: opsGroup, Version: groupResourceVersion, Kind: resourceDef.kind})
	Expect(err).NotTo(HaveOccurred())

	resourceSpecCr, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(testNamespace).Get(context.TODO(), resourceDef.resourceName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	resourceSpecCr.Object["status"] = resourceDef.statusContents

	resourceSpecCr, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(testNamespace).UpdateStatus(context.TODO(), resourceSpecCr, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())

}

func createResourceCr(name string, namespace string, kind string) K8sResourceId {
	return K8sResourceId{
		Name:      name,
		Namespace: namespace,
		Gvk:       schema.GroupVersionKind{Group: opsGroup, Version: groupResourceVersion, Kind: kind},
	}
}

func checkIfResourceDoesNotExist(resourceCrName string, kind string) {
	var gvr schema.GroupVersionResource
	var err error

	gvr, _, err = GetGvrAndAPIResources(schema.GroupVersionKind{Group: opsGroup, Version: groupResourceVersion, Kind: kind})
	Expect(err).NotTo(HaveOccurred())

	_, err = ctk8sclient.GetDynamicK8sClient(ctenv.Cfg).Resource(gvr).Namespace(testNamespace).Get(context.TODO(), resourceCrName, metav1.GetOptions{})
	Expect(k8serrors.IsNotFound(err)).To(BeTrue())
}

func findMissingPorts(portsActual []corev1.ServicePort) []int32 {
	var portMap map[int32]bool
	portMap = makePortMap()

	for _, portDefinition := range portsActual {
		actualPort := portDefinition.Port

		if _, ok := portMap[actualPort]; ok {
			portMap[actualPort] = true
		}
	}

	var missingPorts []int32
	for port, found := range portMap {
		if !found {
			missingPorts = append(missingPorts, port)
		}
	}

	return missingPorts
}

func makePortMap() map[int32]bool {
	var elementMap map[int32]bool
	elementMap = make(map[int32]bool)

	for _, port := range expectedPorts {
		elementMap[port] = false
	}
	return elementMap
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
	}, consulTestDefaultTimeout).Should(HaveOccurred())

}
