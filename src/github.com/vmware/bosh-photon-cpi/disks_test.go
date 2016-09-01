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
	"github.com/vmware/bosh-photon-cpi/cmd"
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"runtime"
)

var _ = Describe("Disk", func() {
	var (
		server *httptest.Server
		ctx    *cpi.Context
		projID string
	)

	BeforeEach(func() {
		server = NewMockServer()
		var runner cmd.Runner
		if runtime.GOOS == "linux" {
			runner = cmd.NewRunner()
		} else {
			runner = &fakeRunner{}
		}

		Activate(true)
		httpClient := &http.Client{Transport: DefaultMockTransport}
		ctx = &cpi.Context{
			Client: ec.NewTestClient(server.URL, nil, httpClient),
			Config: &cpi.Config{
				Photon: &cpi.PhotonConfig{
					Target:    server.URL,
					ProjectID: "fake-project-id",
				},
			},
			Runner: runner,
			Logger: logger.New(),
		}

		projID = ctx.Config.Photon.ProjectID
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CreateDisk", func() {
		It("returns a disk ID", func() {
			createTask := &ec.Task{Operation: "CREATE_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "CREATE_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(200, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_disk": CreateDisk,
			}
			args := []interface{}{2500.0, map[string]interface{}{"disk_flavor": "disk-flavor"}, "fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "create_disk", args))

			Expect(res.Result).Should(Equal(completedTask.Entity.ID))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when size is too small", func() {
			createTask := &ec.Task{Operation: "CREATE_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "CREATE_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(200, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_disk": CreateDisk,
			}
			args := []interface{}{0, map[string]interface{}{"flavor": "disk-flavor"}, "fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "create_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when apife returns a 500", func() {
			createTask := &ec.Task{Operation: "CREATE_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "CREATE_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(500, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_disk": CreateDisk,
			}
			args := []interface{}{2500, map[string]interface{}{"flavor": "disk-flavor"}, "fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "create_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"create_disk": CreateDisk,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "create_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"create_disk": CreateDisk,
			}
			args := []interface{}{"not-an-int"}
			res, err := GetResponse(dispatch(ctx, actions, "create_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})

	Describe("DeleteDisk", func() {
		It("returns nothing", func() {
			deleteTask := &ec.Task{Operation: "DELETE_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "DELETE_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/disks/"+deleteTask.Entity.ID,
				CreateResponder(200, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_disk": DeleteDisk,
			}
			args := []interface{}{"fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when apife returns 404", func() {
			deleteTask := &ec.Task{Operation: "DELETE_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "DELETE_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/disks/"+deleteTask.Entity.ID,
				CreateResponder(404, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_disk": DeleteDisk,
			}
			args := []interface{}{"fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"delete_disk": DeleteDisk,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "delete_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"delete_disk": DeleteDisk,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "delete_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})

	Describe("HasDisk", func() {
		It("returns true for HasDisk when disk exists", func() {
			disk := &ec.PersistentDisk{Flavor: "persistent-disk", ID: "fake-disk-id"}

			RegisterResponder(
				"GET",
				server.URL+"/disks/"+disk.ID,
				CreateResponder(200, ToJson(disk)))

			actions := map[string]cpi.ActionFn{
				"has_disk": HasDisk,
			}
			args := []interface{}{"fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_disk", args))

			Expect(res.Result).Should(Equal(true))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns false for HasDisk when disk does not exists", func() {
			disk := &ec.PersistentDisk{Flavor: "persistent-disk", ID: "fake-disk-id"}

			RegisterResponder(
				"GET",
				server.URL+"/disks/"+disk.ID,
				CreateResponder(404, ToJson(disk)))

			actions := map[string]cpi.ActionFn{
				"has_disk": HasDisk,
			}
			args := []interface{}{"fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_disk", args))

			Expect(res.Result).Should(Equal(false))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error for HasDisk when server returns error", func() {
			disk := &ec.PersistentDisk{Flavor: "persistent-disk", ID: "fake-disk-id"}

			RegisterResponder(
				"GET",
				server.URL+"/disks/"+disk.ID,
				CreateResponder(500, ToJson(disk)))

			actions := map[string]cpi.ActionFn{
				"has_disk": HasDisk,
			}
			args := []interface{}{"fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"has_disk": HasDisk,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "has_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"has_disk": HasDisk,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "has_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		Context("when auth is enabled", func() {
			var (
				ctxAuth    *cpi.Context
			)
			BeforeEach(func() {
				ctxAuth = &cpi.Context{
					Client: ctx.Client,
					Config: &cpi.Config{
						Photon: &cpi.PhotonConfig{
							Target:    ctx.Config.Photon.Target,
							ProjectID: ctx.Config.Photon.ProjectID,
							Username:  "fake_username",
							Password:  "fake_password",
						},
						Agent: ctx.Config.Agent,
					},
					Runner: ctx.Runner,
					Logger: ctx.Logger,
				}
			})

			It("should return false when disk not found", func() {
				disk := &ec.PersistentDisk{Flavor: "persistent-disk", ID: "fake-disk-id"}

				RegisterResponder(
					"GET",
					server.URL+"/disks/"+disk.ID,
					CreateResponder(403, ToJson(disk)))

				actions := map[string]cpi.ActionFn{
					"has_disk": HasDisk,
				}
				args := []interface{}{"fake-disk-id"}
				res, err := GetResponse(dispatch(ctxAuth, actions, "has_disk", args))

				Expect(res.Result).Should(Equal(false))
				Expect(res.Error).Should(BeNil())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.Log).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("GetDisks", func() {
		It("returns a list of VM IDs that a disk is attached to", func() {
			list := &ec.DiskList{
				[]ec.PersistentDisk{
					ec.PersistentDisk{ID: "disk-1", VMs: []string{"vm-1", "vm-2"}},
					ec.PersistentDisk{ID: "disk-2", VMs: []string{"vm-2", "vm-3"}},
					ec.PersistentDisk{ID: "disk-3", VMs: []string{"vm-4", "vm-5"}},
					ec.PersistentDisk{ID: "disk-4", VMs: []string{"vm-2", "vm-4"}},
				},
			}
			// Disks on vm-2
			matchedList := []interface{}{"disk-4", "disk-2", "disk-1"}

			RegisterResponder(
				"GET",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(200, ToJson(list)))

			actions := map[string]cpi.ActionFn{
				"get_disks": GetDisks,
			}
			args := []interface{}{"vm-2"}
			res, err := GetResponse(dispatch(ctx, actions, "get_disks", args))

			Expect(res.Result).Should(ConsistOf(matchedList))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an empty list if no disks are attached to VM", func() {
			list := &ec.DiskList{
				[]ec.PersistentDisk{
					ec.PersistentDisk{ID: "disk-1", VMs: []string{"other-vm"}},
				},
			}
			matchedList := []interface{}{}

			RegisterResponder(
				"GET",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(200, ToJson(list)))

			actions := map[string]cpi.ActionFn{
				"get_disks": GetDisks,
			}
			args := []interface{}{"vm-2"}
			res, err := GetResponse(dispatch(ctx, actions, "get_disks", args))

			Expect(res.Result).Should(ConsistOf(matchedList))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when server returns error", func() {
			list := &ec.DiskList{[]ec.PersistentDisk{}}

			RegisterResponder(
				"GET",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(500, ToJson(list)))

			actions := map[string]cpi.ActionFn{
				"get_disks": GetDisks,
			}
			args := []interface{}{"vm-2"}
			res, err := GetResponse(dispatch(ctx, actions, "get_disks", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
	Describe("AttachDisk", func() {
		It("returns nothing when attach succeeds", func() {
			attachTask := &ec.Task{Operation: "ATTACH_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "ATTACH_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			detachIsoTask := &ec.Task{Operation: "DETACH_ISO", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			detachCompletedTask := &ec.Task{Operation: "DETACH_ISO", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			isoTask := &ec.Task{Operation: "ATTACH_ISO", State: "QUEUED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			isoCompletedTask := &ec.Task{Operation: "ATTACH_ISO", State: "COMPLETED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			env := &cpi.AgentEnv{AgentID: "agent-id", VM: cpi.VMSpec{ID: "fake-vm-id", Name: "fake-vm"}}
			vm := &ec.VM{
				ID:       "fake-vm-id",
				Metadata: map[string]string{"bosh-cpi": GetEnvMetadata(env)},
			}
			metadataTask := &ec.Task{State: "COMPLETED"}

			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/attach_disk",
				CreateResponder(200, ToJson(attachTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/attach_iso",
				CreateResponder(200, ToJson(isoTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/detach_iso",
				CreateResponder(200, ToJson(detachIsoTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/set_metadata",
				CreateResponder(200, ToJson(metadataTask)))
			RegisterResponder(
				"GET",
				server.URL+"/vms/fake-vm-id",
				CreateResponder(200, ToJson(vm)))

			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+attachTask.ID,
				CreateResponder(200, ToJson(completedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+isoTask.ID,
				CreateResponder(200, ToJson(isoCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+detachIsoTask.ID,
				CreateResponder(200, ToJson(detachCompletedTask)))

			actions := map[string]cpi.ActionFn{
				"attach_disk": AttachDisk,
			}
			args := []interface{}{"fake-vm-id", "fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "attach_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when VM not found", func() {
			attachTask := &ec.Task{Operation: "ATTACH_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "ATTACH_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/attach_disk",
				CreateResponder(404, ToJson(attachTask)))

			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+attachTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"attach_disk": AttachDisk,
			}
			args := []interface{}{"fake-vm-id", "fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "attach_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
	Describe("DetachDisk", func() {
		It("returns nothing when detach succeeds", func() {
			attachTask := &ec.Task{Operation: "DETACH_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "DETACH_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			detachIsoTask := &ec.Task{Operation: "DETACH_ISO", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			detachCompletedTask := &ec.Task{Operation: "DETACH_ISO", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			isoTask := &ec.Task{Operation: "ATTACH_ISO", State: "QUEUED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			isoCompletedTask := &ec.Task{Operation: "ATTACH_ISO", State: "COMPLETED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			env := &cpi.AgentEnv{AgentID: "agent-id", VM: cpi.VMSpec{ID: "fake-vm-id", Name: "fake-vm"}}
			vm := &ec.VM{
				ID:       "fake-vm-id",
				Metadata: map[string]string{"bosh-cpi": GetEnvMetadata(env)},
			}
			metadataTask := &ec.Task{State: "COMPLETED"}

			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/detach_disk",
				CreateResponder(200, ToJson(attachTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/attach_iso",
				CreateResponder(200, ToJson(isoTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/detach_iso",
				CreateResponder(200, ToJson(detachIsoTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/set_metadata",
				CreateResponder(200, ToJson(metadataTask)))
			RegisterResponder(
				"GET",
				server.URL+"/vms/fake-vm-id",
				CreateResponder(200, ToJson(vm)))

			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+attachTask.ID,
				CreateResponder(200, ToJson(completedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+isoTask.ID,
				CreateResponder(200, ToJson(isoCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+detachIsoTask.ID,
				CreateResponder(200, ToJson(detachCompletedTask)))

			actions := map[string]cpi.ActionFn{
				"detach_disk": DetachDisk,
			}
			args := []interface{}{"fake-vm-id", "fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "detach_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("returns an error when VM not found", func() {
			attachTask := &ec.Task{Operation: "DETACH_DISK", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}
			completedTask := &ec.Task{Operation: "DETACH_DISK", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-disk-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/detach_disk",
				CreateResponder(404, ToJson(attachTask)))

			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+attachTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"attach_disk": DetachDisk,
			}
			args := []interface{}{"fake-vm-id", "fake-disk-id"}
			res, err := GetResponse(dispatch(ctx, actions, "attach_disk", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
})
