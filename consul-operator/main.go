// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/config"
	"github.com/nokia/industrial-application-framework/application-lib/pkg/handlers"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/licenceexpired"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/monitoring"
	"github.com/nokia/industrial-application-framework/consul-operator/pkg/parameters"
	"github.com/operator-framework/operator-lib/leader"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appdacnokiacomv1alpha1 "github.com/nokia/industrial-application-framework/consul-operator/api/v1alpha1"
	"github.com/nokia/industrial-application-framework/consul-operator/controllers"
	//+kubebuilder:scaffold:imports
)

const (
	configDir = "config/operator"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appdacnokiacomv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8383", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	// Become the leader before proceeding
	// to keep compatibility with previous operator sdk
	err = leader.Become(context.TODO(), "consul-operator-lock")
	if err != nil {
		setupLog.Error(err, "unable to become the leader")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "612bd6e8.app.dac.nokia.com",
		Namespace:              watchNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	var operatorConfig config.OperatorConfig
	operatorConfig, err = config.GetConfiguration(configDir)
	if err != nil {
		setupLog.Error(err, "unable to read configuration")
		os.Exit(1)
	}

	reconciler := controllers.ConsulReconciler{
		Common: handlers.OperatorReconciler{
			Client:        mgr.GetClient(),
			Scheme:        mgr.GetScheme(),
			Configuration: operatorConfig,
			Functions: handlers.ReconcilerHookFunctions{
				CreateAppCr:                   appdacnokiacomv1alpha1.CreateAppInstance,
				CreateAppStatusMonitor:        monitoring.CreateAppStatusMonitor,
				CreateLicenceExpiredHandler:   licenceexpired.CreateLicenseExpiredHandler,
				CheckNetworkParametersChanged: parameters.NetworkParametersChanged,
			},
		},
	}

	if err = controllers.SetupWithManager(mgr, &reconciler); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Consul")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
