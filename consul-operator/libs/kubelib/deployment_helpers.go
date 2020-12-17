package kubelib

import apps "k8s.io/api/apps/v1beta1"

func CreateDeployment(name string) *apps.Deployment {
	dep := &apps.Deployment{}
	dep.Kind = "Deployment"
	dep.APIVersion = "extensions/v1beta1"
	dep.Name = name
	var revLimit int32
	dep.Spec.RevisionHistoryLimit = &revLimit
	var i int32 = 1
	dep.Spec.Replicas = &i
	return dep

}
