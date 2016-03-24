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
	"net/http"

	"github.com/vmware/bosh-photon-cpi/cpi"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
)

type CloudProps struct {
	VMFlavor             string
	DiskFlavor           string
	VMAttachedDiskSizeGB int
}

const (
	VMAttachedDiskSizeGBDefault = 16
	DiskFlavorElement           = "disk_flavor"
	VMFlavorElement             = "vm_flavor"
	VMAttachedDiskSizeGBElement = "vm_attached_disk_size_gb"
)

var ErrCloudPropsValues = errors.New("error in cloud props properties")

func ParseCloudProps(cloudPropsMap map[string]interface{}) (cloudProps CloudProps, err error) {
	var (
		diskOk bool
		vmOk   bool
	)
	cloudProps.DiskFlavor, diskOk = cloudPropsMap[DiskFlavorElement].(string)
	cloudProps.VMFlavor, vmOk = cloudPropsMap[VMFlavorElement].(string)
	cloudProps.VMAttachedDiskSizeGB = VMAttachedDiskSizeGBDefault

	if _, ok := cloudPropsMap[VMAttachedDiskSizeGBElement]; ok {
		cloudProps.VMAttachedDiskSizeGB = int(cloudPropsMap[VMAttachedDiskSizeGBElement].(float64))
	}
	if !diskOk || !vmOk {
		err = ErrCloudPropsValues
	}
	return
}

func CreateVM(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 6 {
		return nil, errors.New("Expected at least 6 arguments")
	}
	agentID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where agent_id should be")
	}
	stemcellCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where stemcell_cid should be")
	}
	cloudPropsMap, ok := args[2].(map[string]interface{})
	if !ok {
		return nil, errors.New("Unexpected argument where cloud_properties should be")
	}

	cloudProps, err := ParseCloudProps(cloudPropsMap)
	if err != nil {
		return nil, err
	}

	networks, ok := args[3].(map[string]interface{})
	if !ok {
		return nil, errors.New("Unexpected argument where networks should be")
	}
	// Ignore args[4] for now, which is disk_cids
	env, ok := args[5].(map[string]interface{})
	if !ok {
		return nil, errors.New("Unexpected argument where env should be")
	}

	ctx.Logger.Infof(
		"CreateVM with agent_id: '%v', stemcell_cid: '%v', cloud_properties: '%v', networks: '%v', env: '%v'",
		agentID, stemcellCID, cloudProps, networks, env)

	ephDiskName := "bosh-ephemeral-disk"
	spec := &ec.VmCreateSpec{
		Name:          "bosh-vm",
		Flavor:        cloudProps.VMFlavor,
		SourceImageID: stemcellCID,
		AttachedDisks: []ec.AttachedDisk{
			ec.AttachedDisk{
				CapacityGB: 50, // Ignored
				Flavor:     cloudProps.DiskFlavor,
				Kind:       "ephemeral-disk",
				Name:       "boot-disk",
				State:      "STARTED",
				BootDisk:   true,
			},
			ec.AttachedDisk{
				CapacityGB: cloudProps.VMAttachedDiskSizeGB,
				Flavor:     cloudProps.DiskFlavor,
				Kind:       "ephemeral-disk",
				Name:       ephDiskName,
				State:      "STARTED",
				BootDisk:   false,
			},
		},
	}
	ctx.Logger.Infof("Creating VM with spec: %#v", spec)
	vmTask, err := ctx.Client.Projects.CreateVM(ctx.Config.Photon.ProjectID, spec)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", vmTask)
	vmTask, err = ctx.Client.Tasks.Wait(vmTask.ID)
	if err != nil {
		return
	}

	// Get disk details of VM
	ctx.Logger.Infof("Getting details of VM: %s", vmTask.Entity.ID)
	vm, err := ctx.Client.VMs.Get(vmTask.Entity.ID)
	if err != nil {
		return
	}
	diskID := ""
	for _, disk := range vm.AttachedDisks {
		if disk.Name == ephDiskName {
			diskID = disk.ID
			break
		}
	}
	if diskID == "" {
		err = cpi.NewBoshError(
			cpi.CloudError, false, "Could not find ID for ephemeral disk of new VM %s", vm.ID)
		return
	}

	// Create agent config
	agentEnv := &cpi.AgentEnv{
		AgentID:  agentID,
		VM:       cpi.VMSpec{Name: vm.Name, ID: vm.ID},
		Networks: networks,
		Env:      env,
		Mbus:     ctx.Config.Agent.Mbus,
		NTP:      ctx.Config.Agent.NTP,
		Disks: map[string]interface{}{
			"ephemeral": map[string]interface{}{
				"id":   diskID,
				"path": "/dev/sdb"},
		},
		Blobstore: cpi.BlobstoreSpec{
			Provider: ctx.Config.Agent.Blobstore.Provider,
			Options:  ctx.Config.Agent.Blobstore.Options,
		},
	}

	// Create and attach agent env ISO file
	err = updateAgentEnv(ctx, vmTask.Entity.ID, agentEnv)
	if err != nil {
		return
	}

	ctx.Logger.Info("Starting VM")
	onTask, err := ctx.Client.VMs.Start(vmTask.Entity.ID)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", onTask)
	onTask, err = ctx.Client.Tasks.Wait(onTask.ID)
	if err != nil {
		return
	}

	return vmTask.Entity.ID, nil
}

func DeleteVM(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}

	ctx.Logger.Infof("Deleting VM: %s", vmCID)

	ctx.Logger.Info("Detaching disks")
	// Detach any attached disks first
	disks, err := ctx.Client.Projects.GetDisks(ctx.Config.Photon.ProjectID, nil)
	if err != nil {
		return
	}
	for _, disk := range disks.Items {
		for _, vmID := range disk.VMs {
			if vmID == vmCID {
				ctx.Logger.Infof("Detaching disk: %s", disk.ID)
				detachOp := &ec.VmDiskOperation{DiskID: disk.ID}
				detachTask, err := ctx.Client.VMs.DetachDisk(vmCID, detachOp)
				if err != nil {
					return nil, err
				}
				ctx.Logger.Infof("Waiting on task: %#v", detachTask)
				detachTask, err = ctx.Client.Tasks.Wait(detachTask.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	ctx.Logger.Info("Stopping VM")
	offTask, err := ctx.Client.VMs.Stop(vmCID)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", offTask)
	offTask, err = ctx.Client.Tasks.Wait(offTask.ID)
	if err != nil {
		return
	}

	ctx.Logger.Info("Deleting VM")
	task, err := ctx.Client.VMs.Delete(vmCID)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", task)
	_, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}

func HasVM(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}

	ctx.Logger.Infof("Determining if VM exists: %s", vmCID)
	_, err = ctx.Client.VMs.Get(vmCID)
	if err != nil {
		apiErr, ok := err.(ec.ApiError)
		if ok && apiErr.HttpStatusCode == http.StatusNotFound {
			return false, nil
		}
		return nil, err
	}
	return true, nil
}

func RestartVM(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}

	ctx.Logger.Infof("Restarting VM: %s", vmCID)
	task, err := ctx.Client.VMs.Restart(vmCID)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", task)
	_, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}
