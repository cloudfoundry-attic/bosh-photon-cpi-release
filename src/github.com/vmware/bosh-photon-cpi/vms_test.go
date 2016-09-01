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
	"net/http"
	"net/http/httptest"
	"runtime"

	"github.com/vmware/bosh-photon-cpi/cmd"
	"github.com/vmware/bosh-photon-cpi/cpi"
	"github.com/vmware/bosh-photon-cpi/logger"
	. "github.com/vmware/bosh-photon-cpi/mocks"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VMs", func() {
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
				Agent: &cpi.AgentConfig{Mbus: "fake-mbus", NTP: []string{"fake-ntp"}},
			},
			Runner: runner,
			Logger: logger.New(),
		}

		projID = ctx.Config.Photon.ProjectID
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("given a ParseCloudProps function", func() {
		var (
			controlDisk                  = "core-100"
			controlVM                    = "core-102"
			controlSize          float64 = 60
			controlCloudPropsMap         = map[string]interface{}{
				DiskFlavorElement:           controlDisk,
				VMFlavorElement:             controlVM,
				VMAttachedDiskSizeGBElement: controlSize,
			}
		)
		Context("when given a valid cloud props map", func() {
			It("then it should set the proper value for DiskFlavor in the response", func() {
				cloudProps, err := ParseCloudProps(controlCloudPropsMap)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cloudProps.DiskFlavor).Should(Equal(controlDisk))
			})

			It("then it should set the proper value for VMFlavor in the response", func() {
				cloudProps, err := ParseCloudProps(controlCloudPropsMap)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cloudProps.VMFlavor).Should(Equal(controlVM))
			})

			It("then it should set the proper value for Attached Disk Size in the response", func() {
				cloudProps, err := ParseCloudProps(controlCloudPropsMap)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cloudProps.VMAttachedDiskSizeGB).Should(Equal(int(controlSize)))
			})

			Context("when given a cloud prop map not containing a `vm_attached_disk_size_gb` element", func() {
				It("then it should use the default value for Attached Disk Size in the response", func() {
					delete(controlCloudPropsMap, VMAttachedDiskSizeGBElement)
					cloudProps, err := ParseCloudProps(controlCloudPropsMap)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cloudProps.VMAttachedDiskSizeGB).Should(Equal(VMAttachedDiskSizeGBDefault))
				})
			})
		})
	})

	Describe("CreateVM", func() {
		It("should return ID of created VM", func() {
			createTask := &ec.Task{Operation: "CREATE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "CREATE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			isoTask := &ec.Task{Operation: "ATTACH_ISO", State: "QUEUED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			isoCompletedTask := &ec.Task{Operation: "ATTACH_ISO", State: "COMPLETED", ID: "fake-iso-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			onTask := &ec.Task{Operation: "START_VM", State: "QUEUED", ID: "fake-on-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			onCompletedTask := &ec.Task{Operation: "START_VM", State: "COMPLETED", ID: "fake-on-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			detachTask := &ec.Task{Operation: "DETACH_ISO", State: "ERROR", ID: "fake-detach-id"}

			vm := &ec.VM{
				ID: createTask.Entity.ID,
				AttachedDisks: []ec.AttachedDisk{
					ec.AttachedDisk{Name: "bosh-ephemeral-disk", ID: "fake-eph-disk-id"},
				},
			}
			metadataTask := &ec.Task{State: "COMPLETED"}

			RegisterResponder(
				"POST",
				server.URL+"/projects/"+projID+"/vms",
				CreateResponder(200, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/vms/"+createTask.Entity.ID,
				CreateResponder(200, ToJson(vm)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+createTask.Entity.ID+"/attach_iso",
				CreateResponder(200, ToJson(isoTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+createTask.Entity.ID+"/detach_iso",
				CreateResponder(200, ToJson(detachTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+createTask.Entity.ID+"/operations",
				CreateResponder(200, ToJson(onTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+createTask.Entity.ID+"/start",
				CreateResponder(200, ToJson(onTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/fake-vm-id/set_metadata",
				CreateResponder(200, ToJson(metadataTask)))

			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+isoTask.ID,
				CreateResponder(200, ToJson(isoCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+onCompletedTask.ID,
				CreateResponder(200, ToJson(onCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+detachTask.ID,
				CreateResponder(200, ToJson(detachTask)))

			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{
				"agent-id",
				"fake-stemcell-id",
				map[string]interface{}{
					"vm_flavor":   "fake-flavor",
					"disk_flavor": "fake-flavor",
				}, // cloud_properties
				map[string]interface{}{}, // networks
				[]interface{}{},          // disk_cids
				map[string]interface{}{}, // environment
			}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(Equal(completedTask.Entity.ID))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when server returns error", func() {
			createTask := &ec.Task{Operation: "CREATE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "CREATE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/projects/"+projID+"/vms",
				CreateResponder(500, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{
				"agent-id",
				"fake-stemcell-id",
				map[string]interface{}{
					"vm_flavor":   "fake-flavor",
					"disk_flavor": "fake-flavor",
				}, // cloud_properties
				map[string]interface{}{}, // networks
				[]string{},               // disk_cids
				map[string]interface{}{}, // environment
			}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when cloud_properties has bad property type", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{
				"agent-id",
				"fake-stemcell-id",
				map[string]interface{}{
					"vm_flavor":   123,
					"disk_flavor": "fake-flavor",
				}, // cloud_properties
				map[string]interface{}{}, // networks
				[]string{},               // disk_cids
				map[string]interface{}{}, // environment
			}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when cloud_properties has no properties", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"agent-id", "fake-stemcell-id", map[string]interface{}{}}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when cloud_properties is missing", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"agent-id", "fake-stemcell-id"}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})

	Describe("DeleteVM", func() {
		It("should return nothing when successful", func() {
			deleteTask := &ec.Task{Operation: "DELETE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "DELETE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			offTask := &ec.Task{Operation: "STOP_VM", State: "QUEUED", ID: "fake-off-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			offCompletedTask := &ec.Task{Operation: "STOP_VM", State: "COMPLETED", ID: "fake-off-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			disks := &ec.DiskList{[]ec.PersistentDisk{
				ec.PersistentDisk{ID: "fake-disk-1", VMs: []string{completedTask.Entity.ID}},
			}}
			detachQueuedTask := &ec.Task{Operation: "DETACH_DISK", State: "QUEUED", ID: "fake-disk-task-1", Entity: ec.Entity{ID: "fake-disk-1"}}
			detachCompletedTask := &ec.Task{Operation: "DETACH_DISK", State: "COMPLETED", ID: "fake-disk-task-1", Entity: ec.Entity{ID: "fake-disk-1"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/vms/"+deleteTask.Entity.ID,
				CreateResponder(200, ToJson(deleteTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+deleteTask.Entity.ID+"/stop",
				CreateResponder(200, ToJson(offTask)))
			RegisterResponder(
				"POST",
				server.URL+"/vms/"+deleteTask.Entity.ID+"/detach_disk",
				CreateResponder(200, ToJson(detachQueuedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+detachCompletedTask.ID,
				CreateResponder(200, ToJson(detachCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+offCompletedTask.ID,
				CreateResponder(200, ToJson(offCompletedTask)))
			RegisterResponder(
				"GET",
				server.URL+"/projects/"+projID+"/disks",
				CreateResponder(200, ToJson(disks)))

			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when VM not found", func() {
			deleteTask := &ec.Task{Operation: "DELETE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "DELETE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/vms/"+deleteTask.Entity.ID,
				CreateResponder(404, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
	Describe("HasVM", func() {
		It("should return true when VM is found", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/vms/"+vm.ID,
				CreateResponder(200, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(Equal(true))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return false when VM not found", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/vms/"+vm.ID,
				CreateResponder(404, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(Equal(false))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when server returns error", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/vms/"+vm.ID,
				CreateResponder(500, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

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

			It("should return false when VM not found", func() {
				vm := &ec.VM{ID: "fake-vm-id"}
				RegisterResponder(
					"GET",
					server.URL+"/vms/"+vm.ID,
					CreateResponder(403, ToJson(vm)))

				actions := map[string]cpi.ActionFn{
					"has_vm": HasVM,
				}
				args := []interface{}{"fake-vm-id"}
				res, err := GetResponse(dispatch(ctxAuth, actions, "has_vm", args))

				Expect(res.Result).Should(Equal(false))
				Expect(res.Error).Should(BeNil())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.Log).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("RestartVM", func() {
		It("should return nothing when successful", func() {
			restartTask := &ec.Task{Operation: "restart_vm", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "restart_vm", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/vms/"+restartTask.Entity.ID+"/restart",
				CreateResponder(200, ToJson(restartTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+restartTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"restart_vm": RestartVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "restart_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when VM not found", func() {
			restartTask := &ec.Task{Operation: "restart_vm", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "restart_vm", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/vms/"+restartTask.Entity.ID+"/operations",
				CreateResponder(404, ToJson(restartTask)))
			RegisterResponder(
				"GET",
				server.URL+"/tasks/"+restartTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"restart_vm": RestartVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "restart_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"restart_vm": RestartVM,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "restart_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"restart_vm": RestartVM,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "restart_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Log).ShouldNot(BeEmpty())
		})
	})
})
