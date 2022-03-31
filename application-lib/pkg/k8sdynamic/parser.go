// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package k8sdynamic

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

const ResourceSeparator = "---"

func splitToIndividualResources(concatenatedResourceList string) []string {
	return strings.Split(concatenatedResourceList, ResourceSeparator)
}

func yamlToUnstructured(yamlStr string) (unstructured.Unstructured, error) {
	jsonContent, err := yaml.ToJSON([]byte(yamlStr))
	if err != nil {
		return unstructured.Unstructured{}, errors.Wrap(err, "failed to convert yaml resource to json")
	}

	var out map[string]interface{}
	err = json.Unmarshal(jsonContent, &out)
	if err != nil {
		return unstructured.Unstructured{}, errors.Wrap(err, "failed to convert json to map struct")
	}

	return unstructured.Unstructured{Object: out}, nil
}
