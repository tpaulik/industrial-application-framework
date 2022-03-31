// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package kubelib

import (
	batchv1 "k8s.io/api/batch/v1"
)

//Create a basic job
func CreateJob(name string) *batchv1.Job {
	j := &batchv1.Job{}
	j.APIVersion = "batch/v1"
	j.Kind = "Job"
	j.Name = name
	j.Spec.Template.Name = name
	j.Spec.Template.Spec.RestartPolicy = "Never"
	return j
}
