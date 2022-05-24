// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

const OperatorConfigFilename = "operatorconfig.yaml"

type OperatorConfig struct {
	ApplicationName             string         `yaml:"applicationName"`
	Namespace                   string         `yaml:"namespace"`
	SourceDeploymentPath        string         `yaml:"sourceDeploymentPath"`
	RuntimeDeploymentPath       string         `yaml:"runtimeDeploymentPath"`
	AppDeploymentDirName        string         `yaml:"appDeploymentDirName"`
	RuntimeResReqPath           string         `yaml:"runtimeResReqPath"`
	ResReqDirName               string         `yaml:"resReqDirName"`
	KubernetesAppDeploymentName string         `yaml:"kubernetesAppDeploymentName"`
	AppPnaName                  string         `yaml:"appPnaName"`
	Template                    TemplateConfig `yaml:"templater"`
}

func (in *OperatorConfig) GetAppDeploymentSourcePath() string {
	return in.SourceDeploymentPath + "/" + in.AppDeploymentDirName
}

func (in *OperatorConfig) GetResourceRequestSourcePath() string {
	return in.SourceDeploymentPath + "/" + in.ResReqDirName
}

type TemplateConfig struct {
	LeftDelimiter  string `yaml:"leftDelimiter"`
	RightDelimiter string `yaml:"rightDelimiter"`
}

func GetConfiguration(configDir string) (OperatorConfig, error) {
	operatorConfigYaml := configDir + "/" + OperatorConfigFilename

	config := OperatorConfig{}
	var err error = nil

	data, err := os.ReadFile(operatorConfigYaml)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal([]byte(data), &config)
	return config, err
}
