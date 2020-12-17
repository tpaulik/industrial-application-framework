package kubelib

import k8v1 "k8s.io/api/core/v1"

func AddPodEmptyVolume(p *k8v1.PodSpec, name string) {

	v := k8v1.Volume{Name: name}
	v.EmptyDir = &k8v1.EmptyDirVolumeSource{}
	p.Volumes = append(p.Volumes, v)

}

func AddPodHostVolume(p *k8v1.PodSpec, name string, hostpath string) {
	v := k8v1.Volume{}
	v.Name = name
	v.HostPath = &k8v1.HostPathVolumeSource{}
	v.HostPath.Path = hostpath
	p.Volumes = append(p.Volumes, v)

}

//Adds a configMap volume to a pod
func AddPodConfigVolume(p *k8v1.PodSpec, volName string, configMap string) {
	v := k8v1.Volume{}
	v.Name = volName
	v.ConfigMap = &k8v1.ConfigMapVolumeSource{}
	v.ConfigMap.Name = configMap
	p.Volumes = append(p.Volumes, v)

}

//AddContainer adds container to a pod
func AddContainer(p *k8v1.PodSpec, c *k8v1.Container) {
	p.Containers = append(p.Containers, *c)
}
