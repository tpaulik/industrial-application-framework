module github.com/nokia/industrial-application-framework/componenttest-lib

go 1.17

require (
	github.com/onsi/ginkgo/v2 v2.1.4
    github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	go.etcd.io/etcd/client/v3 v3.5.0
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.24.1
    k8s.io/client-go v0.24.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.12.1
)
