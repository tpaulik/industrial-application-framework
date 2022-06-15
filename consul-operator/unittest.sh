#!/bin/bash
# Copyright 2022 Nokia
# Licensed under the BSD 3-Clause License.
# SPDX-License-Identifier: BSD-3-Clause

set -eo pipefail
COVERPKG=$(go list ./... | grep -v componenttest | tr '\n' ',' | sed 's/,//')
go test ./componenttest -v -coverprofile=coverage_ct.out -race -covermode=atomic -json -ginkgo.v -coverpkg=$COVERPKG 2>&1 | tee test-report.json
go test $(go list ./... | grep -v /componenttest) -v -coverprofile=coverage_ut.out -race -covermode=atomic -json -coverpkg=./pkg/...  2>&1 | tee -a test-report.json
cat coverage_ct.out > coverage.out
if [[  $(cat coverage_ut.out | grep -v "mode:")  ]]; then
 cat coverage_ut.out | grep -v "mode:" >> coverage.out
fi
