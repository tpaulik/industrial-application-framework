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
	ApplicationName     string         `yaml:"applicationName"`
	Namespace           string         `yaml:"namespace"`
	DeploymentSourceDir string         `yaml:"deploymentSourceDir"`
	DeploymentDir       string         `yaml:"deploymentDir"`
	DeploymentDirName   string         `yaml:"deploymentDirName"`
	ResReqDir           string         `yaml:"resReqDir"`
	ResReqDirName       string         `yaml:"resReqDirName"`
	ServiceName         string         `yaml:"serviceName"`
	DeploymentName      string         `yaml:"deploymentName"`
	AppPnaName          string         `yaml:"appPnaName"`
	Template            TemplateConfig `yaml:"templater"`
}

func (in *OperatorConfig) GetAppDeploymentSourceDir() string {
	return in.DeploymentSourceDir + "/" + in.DeploymentDirName
}

func (in *OperatorConfig) GetResourceRequestSourceDir() string {
	return in.DeploymentSourceDir + "/" + in.ResReqDirName
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
