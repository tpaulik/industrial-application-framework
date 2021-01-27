module github.com/nokia/industrial-application-framework/consul-operator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v0.0.0-20191119172530-79f836b90111
	github.com/nokia/industrial-application-framework/alarmlogger v0.0.0-20210121171643-003e1aa52c6d
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
