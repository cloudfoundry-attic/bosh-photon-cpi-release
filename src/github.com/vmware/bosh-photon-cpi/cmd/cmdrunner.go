// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package cmd

import (
	"os/exec"
)

type Runner interface {
	Run(name string, args ...string) ([]byte, error)
}

type defaultRunner struct {
}

func NewRunner() Runner {
	return &defaultRunner{}
}

func (_ defaultRunner) Run(name string, args ...string) (out []byte, err error) {
	cmd := exec.Command(name, args...)
	out, err = cmd.CombinedOutput()
	return
}
