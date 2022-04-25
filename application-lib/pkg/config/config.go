// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

const operatorConfigFilename = "operatorconfig.yaml"

type OperatorConfig struct {
	ApplicationName   string         `yaml:"applicationName"`
	DeploymentDir     string         `yaml:"deploymentDir"`
	DeploymentDirName string         `yaml:"deploymentDirName"`
	ResReqDir         string         `yaml:"resReqDir"`
	ResReqDirName     string         `yaml:"resReqDirName"`
	ServiceName       string         `yaml:"serviceName"`
	DeploymentName    string         `yaml:"deploymentName"`
	AppPnaName        string         `yaml:"appPnaName"`
	Template          TemplateConfig `yaml:"templater"`
}

type TemplateConfig struct {
	LeftDelimiter  string `yaml:"leftDelimiter"`
	RightDelimiter string `yaml:"rightDelimiter"`
}

func GetConfiguration(configDir string) (OperatorConfig, error) {
	operatorConfigYaml := configDir + "/" + operatorConfigFilename

	config := OperatorConfig{}
	var err error = nil

	data, err := os.ReadFile(operatorConfigYaml)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal([]byte(data), &config)
	return config, err
}
