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
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VM metadata", func() {
	var (
		ctx *cpi.Context
	)

	BeforeEach(func() {
		ctx = &cpi.Context{
			Logger: logger.New(),
		}
	})

	It("set_vm_metadata is a nop", func() {
		actions := map[string]cpi.ActionFn{
			"set_vm_metadata": SetVmMetadata,
		}
		res, err := GetResponse(dispatch(ctx, actions, "set_vm_metadata", nil))
		Expect(res.Result).To(BeNil())
		Expect(err).To(BeNil())
		Expect(res.Log).ShouldNot(BeEmpty())
	})
})
