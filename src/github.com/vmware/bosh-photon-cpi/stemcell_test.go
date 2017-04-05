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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Stemcell", func() {
	var (
		server *httptest.Server
		ctx    *cpi.Context
	)

	BeforeEach(func() {
		server = NewMockServer()

		Activate(true)
		httpClient := &http.Client{Transport: DefaultMockTransport}
		ctx = &cpi.Context{
			Client: ec.NewTestClient(server.URL, nil, httpClient),
			Logger: logger.New(),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Create", func() {
		It("returns a stemcell ID for Create", func() {
			createTask := &ec.Task{Operation: "CREATE_IMAGE", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}
			completedTask := &ec.Task{Operation: "CREATE_IMAGE", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}

			RegisterResponder(
				"POST",
				server.URL+rootUrl+"/images",
				CreateResponder(200, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+rootUrl+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_stemcell": CreateStemcell,
			}
			args := []interface{}{"./testdata/image"}
			res, err := GetResponse(dispatch(ctx, actions, "create_stemcell", args))

			Expect(res.Result).Should(Equal(completedTask.Entity.ID))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})

		It("returns an error when APIfe returns a 500", func() {
			createTask := &ec.Task{Operation: "CREATE_IMAGE", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}
			completedTask := &ec.Task{Operation: "CREATE_IMAGE", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}

			RegisterResponder(
				"POST",
				server.URL+rootUrl+"/images",
				CreateResponder(500, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+rootUrl+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_stemcell": CreateStemcell,
			}
			args := []interface{}{"./testdata/tty_tiny.ova"}
			res, err := GetResponse(dispatch(ctx, actions, "create_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})

		It("returns an error when stemcell file does not exist", func() {
			actions := map[string]cpi.ActionFn{
				"create_stemcell": CreateStemcell,
			}
			args := []interface{}{"a-file-that-does-not-exist"}
			res, err := GetResponse(dispatch(ctx, actions, "create_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"create_stemcell": CreateStemcell,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "create_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"create_stemcell": CreateStemcell,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "create_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})

	Describe("Delete", func() {
		It("returns nothing for stemcell delete", func() {
			deleteTask := &ec.Task{Operation: "DELETE_IMAGE", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}
			completedTask := &ec.Task{Operation: "DELETE_IMAGE", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+rootUrl+"/images/"+deleteTask.Entity.ID,
				CreateResponder(200, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+rootUrl+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_stemcell": DeleteStemcell,
			}
			args := []interface{}{deleteTask.Entity.ID}
			res, err := GetResponse(dispatch(ctx, actions, "delete_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error for missing stemcell delete", func() {
			deleteTask := &ec.Task{Operation: "DELETE_IMAGE", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}
			completedTask := &ec.Task{Operation: "DELETE_IMAGE", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-image-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+rootUrl+"/images/"+deleteTask.Entity.ID,
				CreateResponder(404, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+rootUrl+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_stemcell": DeleteStemcell,
			}
			args := []interface{}{deleteTask.Entity.ID}
			res, err := GetResponse(dispatch(ctx, actions, "delete_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"delete_stemcell": DeleteStemcell,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "delete_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"delete_stemcell": DeleteStemcell,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "delete_stemcell", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
})
