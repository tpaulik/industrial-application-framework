// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package componenttest

import (
	ctenv "github.com/nokia/industrial-application-framework/componenttest-lib/pkg/env"
	appdacnokiacomv1alpha1 "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/controllers"
	"github.com/nokia/industrial-application-framework/consul-operator/libs/kubelib"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/k8sdynamic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"testing"
)

const ItBinaryRelativePath = "/../componenttest/resources"

var _ = BeforeSuite(func() {
}, 60)

var _ = AfterSuite(func() {
}, 60)

var ourScheme = k8sruntime.NewScheme()

func init() {

	utilruntime.Must(clientgoscheme.AddToScheme(ourScheme))

	utilruntime.Must(appdacnokiacomv1alpha1.AddToScheme(ourScheme))
	//+kubebuilder:scaffold:scheme
}

func getTestBinaryPath(testBinariesRelativePath string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return basepath + testBinariesRelativePath
}

func CustomTearUp() {
	var err error
	k8sdynamic.Config = ctenv.Cfg
	kubelib.Config = ctenv.Cfg

	k8sManager, err := ctrl.NewManager(ctenv.Cfg, ctrl.Options{
		Scheme: ourScheme,
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient = k8sManager.GetClient()

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = (&controllers.ConsulReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	By("Starting the Operator")
	go func() {
		defer GinkgoRecover()
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).NotTo(HaveOccurred())
	}()
}

func CustomTearDown() {
	ctenv.ResetEtcd()
}

func TestConsulOperator(t *testing.T) {
	RegisterFailHandler(Fail)

	ctenv.TearUpTestEnv(getTestBinaryPath(ItBinaryRelativePath))

	CustomTearUp()

	RunSpecsWithDefaultAndCustomReporters(t, "Monitoring Operator Component Test Suite", []Reporter{printer.NewlineReporter{}})

	CustomTearDown()
	ctenv.TearDownTestEnv()
}
