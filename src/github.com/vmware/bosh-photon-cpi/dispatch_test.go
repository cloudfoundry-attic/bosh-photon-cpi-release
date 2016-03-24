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
	"errors"
	"fmt"
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("Dispatch", func() {
	var (
		ctx        *cpi.Context
		configPath string
	)

	BeforeEach(func() {
		ctx = &cpi.Context{
			Logger: logger.New(),
		}
	})

	AfterEach(func() {
		if configPath != "" {
			os.Remove(configPath)
		}
	})

	It("returns a valid bosh JSON response given valid arguments", func() {
		actions := map[string]cpi.ActionFn{
			"create_vm": createVM,
		}
		args := []interface{}{"fake-agent-id"}
		res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

		Expect(res.Result).Should(Equal("fake-vm-id"))
		Expect(res.Error).Should(BeNil())
		Expect(err).ShouldNot(HaveOccurred())
	})
	It("returns a valid bosh JSON error when given an invalid argument", func() {
		actions := map[string]cpi.ActionFn{
			"create_vm": createVM,
		}
		args := []interface{}{5}
		res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

		Expect(res.Error).ShouldNot(BeNil())
		Expect(res.Error.Type).Should(Equal(cpi.CpiError))
		Expect(err).ShouldNot(HaveOccurred())
	})
	It("returns a valid bosh JSON error when function errors", func() {
		actions := map[string]cpi.ActionFn{
			"create_vm": createVmError,
		}
		args := []interface{}{"fake-agent-id"}
		res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

		Expect(res.Error).ShouldNot(BeNil())
		Expect(res.Error.Type).Should(Equal(cpi.CpiError))
		Expect(err).ShouldNot(HaveOccurred())
	})
	It("returns a valid bosh JSON error when function panics", func() {
		actions := map[string]cpi.ActionFn{
			"create_vm": createVmPanic,
		}
		args := []interface{}{"fake-agent-id"}
		res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

		Expect(res.Error).ShouldNot(BeNil())
		Expect(res.Error.Type).Should(Equal(cpi.CpiError))
		Expect(err).ShouldNot(HaveOccurred())
	})
	It("returns a valid bosh JSON error when method not implemented", func() {
		actions := map[string]cpi.ActionFn{}
		args := []interface{}{"fake-agent-id"}
		res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

		Expect(res.Error).ShouldNot(BeNil())
		Expect(res.Error.Type).Should(Equal(cpi.NotSupportedError))
		Expect(err).ShouldNot(HaveOccurred())
	})
	It("loads JSON config correctly", func() {
		configFile, err := ioutil.TempFile("", "bosh-photon-cpi-config")
		if err != nil {
			panic(err)
		}
		configPath = configFile.Name()
		jsonConfig := `{"photon":{"Target":"http://none:123"}}`
		configFile.WriteString(jsonConfig)

		context, err := loadConfig(configPath)
		expectedURL := fmt.Sprintf("http://%s:%d", "none", 123)
		Expect(context.Client.Endpoint).Should(Equal(expectedURL))
		Expect(err).Should(BeNil())
	})
})

func createVM(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	_, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where agent_id should be")
	}
	return "fake-vm-id", nil
}

func createVmError(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	return nil, errors.New("error occurred")
}

func createVmPanic(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	panic("oh no!")
}
