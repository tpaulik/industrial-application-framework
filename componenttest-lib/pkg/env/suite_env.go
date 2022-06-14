/*
Copyright 2020 Nokia
Licensed under the BSD 3-Clause License.
SPDX-License-Identifier: BSD-3-Clause
*/

package env

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/cache"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nokia/industrial-application-framework/componenttest-lib/pkg/k8sclient"
	"github.com/nokia/industrial-application-framework/componenttest-lib/pkg/nsdeleter"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	KubeApiServerEnvVariable           = "TEST_ASSET_KUBE_APISERVER"
	EtcdEnvVariable                    = "TEST_ASSET_ETCD"
	KubebuilderControlPlaneStopTimeout = "KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT"
	KubeApiServerBinaryName            = "kube-apiserver"
	EtcdBinaryName                     = "etcd"
	defaultKubebuilderPath             = "/usr/local/kubebuilder/bin"
)

type LocalConfig struct {
	CertDir    string
	KubeConfig string
}

var Cfg *rest.Config
var LocalCfg LocalConfig
var k8sClient client.Client
var testenv *envtest.Environment
var CrdPathsToAdd []string
var namespaceControllerStopper chan struct{}

var log = logf.Log.WithName("suite_env")

func TearUpTestEnv(testBinariesPath string, crdPaths ...string) {
	var err error

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	_, err = os.Stat(defaultKubebuilderPath + "/" + KubeApiServerBinaryName)
	if os.IsNotExist(err) {
		_ = os.Setenv(KubeApiServerEnvVariable, testBinariesPath+"/"+KubeApiServerBinaryName)
	}

	_, err = os.Stat(defaultKubebuilderPath + "/" + EtcdBinaryName)
	if os.IsNotExist(err) {
		_ = os.Setenv(EtcdEnvVariable, testBinariesPath+"/"+EtcdBinaryName)
	}

	if v := os.Getenv(KubebuilderControlPlaneStopTimeout); v == "" {
		os.Setenv(KubebuilderControlPlaneStopTimeout, "120s")
	}

	log.Info("bootstrapping test environment")
	defaultCRDPaths := []string{
		filepath.Join("..", "config", "crd", "bases"),
		filepath.Join(".", "crds"),
	}
	CrdPathsToAdd = append(defaultCRDPaths, crdPaths...)
	testenv = &envtest.Environment{
		CRDDirectoryPaths:     CrdPathsToAdd,
		ErrorIfCRDPathMissing: false,
	}

	Cfg, err = testenv.Start()
	if err != nil {
		panic(err)
	}

	LocalCfg = LocalConfig{CertDir: testenv.ControlPlane.APIServer.CertDir}
	LocalCfg.KubeConfig, err = getFilenameFromDirectoryByExtension(LocalCfg.CertDir, "kubecfg")
	if err != nil {
		panic(err)
	}

	startNamespaceController()
}

func startNamespaceController() {
	client := k8sclient.GetK8sClient(Cfg)
	metadataClient, _ := metadata.NewForConfig(Cfg)
	discoverResourcesFn := client.Discovery().ServerPreferredNamespacedResources
	nsDeleter := nsdeleter.NewNamespacedResourcesDeleter(client.CoreV1().Namespaces(), metadataClient, client.CoreV1(), discoverResourcesFn, v1.FinalizerKubernetes)

	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Namespaces().Informer()
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldNs := oldObj.(*v1.Namespace)
				newNs := newObj.(*v1.Namespace)

				if oldNs.DeletionTimestamp == nil && newNs.DeletionTimestamp != nil {
					//delete all of the resources in the namespace
					log.Info("Deleting namespace", "name:", newNs.Name)
					for true {
						err := nsDeleter.Delete(newNs.Name)
						if _, isResourcesRemainingError := err.(*nsdeleter.ResourcesRemainingError); err != nil && !isResourcesRemainingError {
							log.Error(err, "Failed to delete all of the resources from the namespace")
							return
						}
						if err == nil {
							break
						}
						log.Info("There are remaining items in the namespace, try again to delete")
					}
					log.Info("Namespace deleted", "name:", newNs.Name)
				}
			},
		},
	)

	namespaceControllerStopper = make(chan struct{})
	go informer.Run(namespaceControllerStopper)
}

func ResetEtcd() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{testenv.ControlPlane.Etcd.URL.String()},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = cli.Delete(context.Background(), "/", clientv3.WithFromKey())
	Expect(err).NotTo(HaveOccurred())
}

func ResetEnv() {
	ResetEtcd()
	err := InstallCRDs()
	if err != nil {
		panic("Failed to install CRDs")
	}
}

func InstallCRDs() error {
	_, err := envtest.InstallCRDs(Cfg, envtest.CRDInstallOptions{
		Paths: testenv.CRDDirectoryPaths,
		CRDs:  testenv.CRDs,
	})
	if err != nil {
		switch err.(type) {
		case *k8sapierrors.StatusError:
			serr := err.(*k8sapierrors.StatusError)
			if 409 == serr.ErrStatus.Code {
				return nil
			}
		default:
			return err
		}
	}
	return nil
}

func TearDownTestEnv() {
	close(namespaceControllerStopper)
	log.Info("tearing down the test environment")
	err := testenv.Stop()
	Expect(err).NotTo(HaveOccurred())
}

func getFilenameFromDirectoryByExtension(directoryName string, extension string) (string, error) {

	files, err := ioutil.ReadDir(directoryName)
	if err != nil {
		return "", err
	}

	var result string
	for _, file := range files {
		stringSlice := strings.Split(file.Name(), ".")

		if len(stringSlice) > 1 {
			if stringSlice[len(stringSlice)-1] == extension {
				result = file.Name()
			}
		}
	}

	return directoryName + "/" + result, nil
}
