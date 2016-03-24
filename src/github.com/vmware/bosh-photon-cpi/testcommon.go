// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package main

import (
	"strings"
)

// Mock cmd.Runner implementation. The cmds map matches command prefixes to
// command output. E.g. if you want to mock out a call to "ls" and "file",
// pass in something like map[string]string{"ls": "stdout for ls", "file": "stdout for file"}
// When Run is called it will return output for the first key that is a substring of the
// name argument.
type fakeRunner struct {
	cmds map[string]string
}

func (r *fakeRunner) Run(name string, args ...string) (out []byte, err error) {
	for key := range r.cmds {
		if strings.HasPrefix(name, key) {
			return []byte(r.cmds[key]), nil
		}
	}
	return
}
