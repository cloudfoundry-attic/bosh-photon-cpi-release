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
	"github.com/vmware/bosh-photon-cpi/cmd"
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	"net/http"
	"net/http/httptest"
	"os"
)

var _ = Describe("AgentEnv", func() {
	var (
		ctx    *cpi.Context
		env    *cpi.AgentEnv
		runner cmd.Runner
		server *httptest.Server
	)

	BeforeEach(func() {
		server = NewMockServer()
		runner = cmd.NewRunner()
		httpClient := &http.Client{Transport: DefaultMockTransport}
		ctx = &cpi.Context{
			Client: ec.NewTestClient(server.URL, nil, httpClient),
			Config: &cpi.Config{
				Photon: &cpi.PhotonConfig{
					Target:    server.URL,
					ProjectID: "fake-project-id",
				},
				Agent: &cpi.AgentConfig{Mbus: "fake-mbus", NTP: []string{"fake-ntp"}},
			},
			Runner: runner,
			Logger: logger.New(),
		}
		env = &cpi.AgentEnv{AgentID: "agent-id", VM: cpi.VMSpec{Name: "vm-name", ID: "vm-id"}}
	})

	It("Successfully creates an ISO", func() {
		iso, err := createEnvISO(env, runner)
		defer os.Remove(iso)

		Expect(err).Should(BeNil(), "Test requires mkisofs, install with 'brew install cdrtools' on Mac")

		// Verify we have produced a valid ISO by checking the output of "file <iso>"
		out, err := runner.Run("file", iso)
		outStr := string(out[:])

		Expect(err).Should(BeNil())
		Expect(outStr).Should(ContainSubstring("ISO 9660 CD-ROM"))
	})

	Describe("Metadata", func() {
		It("successfully puts and gets agent env data", func() {
			vmID := "fake-vm-id"
			metadataTask := &ec.Task{State: "COMPLETED"}
			vm := &ec.VM{
				ID:       vmID,
				Metadata: map[string]string{"bosh-cpi": GetEnvMetadata(env)},
			}

			RegisterResponder(
				"POST",
				server.URL+rootUrl+"/vms/"+vmID+"/set_metadata",
				CreateResponder(200, ToJson(metadataTask)))
			RegisterResponder(
				"GET",
				server.URL+rootUrl+"/vms/"+vmID,
				CreateResponder(200, ToJson(vm)))

			err := putAgentEnvMetadata(ctx, vmID, env)
			Expect(err).ToNot(HaveOccurred())

			env2, err := getAgentEnvMetadata(ctx, vmID)
			Expect(err).ToNot(HaveOccurred())
			Expect(env2).Should(Equal(env))
		})
	})
})
