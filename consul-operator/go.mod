module github.com/nokia/industrial-application-framework/consul-operator

go 1.17

require (
	github.com/nokia/industrial-application-framework/alarmlogger v0.0.0-20210824095151-771352d42ef7
	github.com/nokia/industrial-application-framework/application-lib v0.0.0-20220613075302-5ad3c9c3c243
	github.com/nokia/industrial-application-framework/componenttest-lib v0.0.0-20220302154657-2b2f359ea42d
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	github.com/operator-framework/operator-lib v0.6.0
	k8s.io/api v0.24.1
	k8s.io/apimachinery v0.24.1
	k8s.io/client-go v0.24.1
	sigs.k8s.io/controller-runtime v0.12.1
)
