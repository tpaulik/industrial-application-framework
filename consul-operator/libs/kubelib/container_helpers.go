package kubelib

import k8v1 "k8s.io/api/core/v1"

//Adds a container volume mount to a container
func AddContainerVolume(c *k8v1.Container, name string, mountpath string) {

	v := k8v1.VolumeMount{}
	v.Name = name
	v.MountPath = mountpath
	c.VolumeMounts = append(c.VolumeMounts, v)

}

//Add env var to container
func AddEnvVar(c *k8v1.Container, key string, value string) {
	e := k8v1.EnvVar{}
	e.Name = key
	e.Value = value
	c.Env = append(c.Env, e)
}

//Adds env variable from configMap
func AddConfigEnvVar(c *k8v1.Container, envName string, selector string, key string) {
	e := k8v1.EnvVar{}
	e.Name = envName
	e.ValueFrom = &k8v1.EnvVarSource{}

	e.ValueFrom.ConfigMapKeyRef = &k8v1.ConfigMapKeySelector{}
	e.ValueFrom.ConfigMapKeyRef.Name = selector
	e.ValueFrom.ConfigMapKeyRef.Key = key
	c.Env = append(c.Env, e)

}
