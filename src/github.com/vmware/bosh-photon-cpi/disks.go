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
	"github.com/vmware/bosh-photon-cpi/cpi"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	"math"
	"net/http"
)

func CreateDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 3 {
		return nil, errors.New("Expected at least 3 arguments")
	}
	disk_size, ok := args[0].(float64)
	if !ok {
		return nil, errors.New("Unexpected argument where size should be")
	}
	size := toGB(disk_size)
	if size < 1 {
		return nil, errors.New("Must provide a size in MiB that rounds up to at least 1 GiB for photon")
	}
	cloudProps, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("Unexpected argument where cloud_properties should be")
	}
	flavor, ok := cloudProps["disk_flavor"].(string)
	if !ok {
		return nil, errors.New("Property 'disk_flavor' on cloud_properties is not a string")
	}
	vmCID, ok := args[2].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}

	ctx.Logger.Infof(
		"CreateDisk with disk_size: '%v' (rounded to '%v' GiB), cloud_properties: '%v', flavor: '%s', vm_cid: '%s'",
		disk_size, size, cloudProps, flavor, vmCID)

	diskSpec := &ec.DiskCreateSpec{
		Flavor:     flavor,
		Kind:       "persistent-disk",
		CapacityGB: size,
		Name:       "disk-for-vm-" + vmCID,
	}

	ctx.Logger.Infof("Creating disk with spec: %#v", diskSpec)
	task, err := ctx.Client.Projects.CreateDisk(ctx.Config.Photon.ProjectID, diskSpec)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", task)
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return task.Entity.ID, nil
}

func DeleteDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	diskCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}

	ctx.Logger.Infof("DeleteDisk with disk_cid: '%s'", diskCID)

	ctx.Logger.Info("Deleting disk")
	task, err := ctx.Client.Disks.Delete(diskCID)
	if err != nil {
		return
	}

	ctx.Logger.Infof("Waiting on task: %#v", task)
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}

func HasDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	diskCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}

	ctx.Logger.Infof("HasDisk with disk_cid: '%s'", diskCID)

	_, err = ctx.Client.Disks.Get(diskCID)
	if err != nil {
		apiErr, ok := err.(ec.ApiError)
		if ok && apiErr.HttpStatusCode == http.StatusNotFound {
			return false, nil
		}
		return nil, err
	}
	return true, nil
}

func GetDisks(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vim_cid should be")
	}

	ctx.Logger.Infof("GetDisks with vm_cid: '%s'", vmCID)

	disks, err := ctx.Client.Projects.GetDisks(ctx.Config.Photon.ProjectID, nil)
	if err != nil {
		return
	}

	res := []string{}
	for _, disk := range disks.Items {
		for _, vm := range disk.VMs {
			if vm == vmCID {
				res = append(res, disk.ID)
			}
		}
	}
	return res, nil
}

func AttachDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 2 {
		return nil, errors.New("Expected at least 2 arguments")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}
	diskCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}

	ctx.Logger.Infof("AttachDisk with vm_cid: '%s', disk_cid: '%s'", vmCID, diskCID)

	ctx.Logger.Info("Attaching disk")
	op := &ec.VmDiskOperation{DiskID: diskCID}
	task, err := ctx.Client.VMs.AttachDisk(vmCID, op)
	if err != nil {
		return
	}

	ctx.Logger.Infof("Waiting on task: %#v", task)
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}

	ctx.Logger.Info("Getting metadata for VM")
	// Get agent env config from VM metadata and update disk ID
	env, err := getAgentEnvMetadata(ctx, vmCID)
	if err != nil {
		return
	}
	if env.Disks == nil {
		env.Disks = map[string]interface{}{}
	}
	persistent := "persistent"
	if _, ok := env.Disks[persistent]; !ok {
		env.Disks[persistent] = map[string]interface{}{}
	}
	diskMap, ok := env.Disks[persistent].(map[string]interface{})
	if !ok {
		return nil, errors.New("Unexpected type found in VM metadata")
	}
	// Agent expects a mapping of disk_cid to the ID that gets used by the agent
	// to resolve the path to the device. In our case, it is the same ID as disk_cid.
	diskMap[diskCID] = map[string]interface{}{
		"id":   diskCID,
		"path": "",
	}

	err = updateAgentEnv(ctx, vmCID, env)
	if err != nil {
		return
	}

	return nil, nil
}

func DetachDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 2 {
		return nil, errors.New("Expected at least 2 arguments")
	}
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}
	diskCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}

	ctx.Logger.Infof("DetachDisk with vm_cid: '%s', disk_cid: '%s'", vmCID, diskCID)

	ctx.Logger.Info("Detaching disk")
	op := &ec.VmDiskOperation{DiskID: diskCID}
	task, err := ctx.Client.VMs.DetachDisk(vmCID, op)
	if err != nil {
		return
	}

	ctx.Logger.Infof("Waiting on task: %#v", task)
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}

	// Get agent env config from VM metadata and remove disk ID
	ctx.Logger.Info("Getting metadata for VM")
	env, err := getAgentEnvMetadata(ctx, vmCID)
	if err != nil {
		return
	}
	persistent := "persistent"
	if diskMap, ok := env.Disks[persistent].(map[string]interface{}); ok {
		delete(diskMap, diskCID)
	}

	err = updateAgentEnv(ctx, vmCID, env)
	if err != nil {
		return
	}

	return nil, nil
}

func toGB(mb float64) int {
	return int(math.Ceil(mb / 1000.0))
}
