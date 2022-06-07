module github.com/nokia/industrial-application-framework/consul-operator

go 1.17

require (
	github.com/nokia/industrial-application-framework/alarmlogger v0.0.0-20210824095151-771352d42ef7
	github.com/nokia/industrial-application-framework/application-lib v0.0.0-20220524094513-0a2f59f2c825
	github.com/nokia/industrial-application-framework/componenttest-lib v0.0.0-20220302154657-2b2f359ea42d
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/operator-framework/operator-lib v0.6.0 //1
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	sigs.k8s.io/controller-runtime v0.12.1
)
