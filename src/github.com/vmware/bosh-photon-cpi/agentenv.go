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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vmware/bosh-photon-cpi/cmd"
	"github.com/vmware/bosh-photon-cpi/cpi"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	"io/ioutil"
	"os"
	p "path"
)

const metadataKey = "bosh-cpi"

func getAgentEnvMetadata(ctx *cpi.Context, vmID string) (res *cpi.AgentEnv, err error) {
	vm, err := ctx.Client.VMs.Get(vmID)
	if err != nil {
		return
	}
	metadata, ok := vm.Metadata[metadataKey]
	if !ok {
		err = fmt.Errorf("No metadata found with key '%s' for vm ID '%s'", metadataKey, vmID)
	}
	res = &cpi.AgentEnv{}
	err = json.Unmarshal([]byte(metadata), res)
	return
}

func putAgentEnvMetadata(ctx *cpi.Context, vmID string, env *cpi.AgentEnv) (err error) {
	envJson, err := json.Marshal(env)
	if err != nil {
		return
	}
	envString := string(envJson[:])
	metadata := &ec.VmMetadata{map[string]string{metadataKey: envString}}
	// Task returns instantly for SetMetadata
	_, err = ctx.Client.VMs.SetMetadata(vmID, metadata)
	return
}

func createEnvISO(env *cpi.AgentEnv, runner cmd.Runner) (path string, err error) {
	json, err := json.Marshal(env)
	if err != nil {
		return
	}
	envDir, err := ioutil.TempDir("", "agent-iso-dir")
	if err != nil {
		return
	}
	// Name of the environment JSON file should be "env" to fit ISO 9660 8.3 filename scheme
	envFile, err := os.Create(p.Join(envDir, "env"))
	if err != nil {
		return
	}
	_, err = envFile.Write(json)
	if err != nil {
		return
	}
	err = envFile.Close()
	if err != nil {
		return
	}

	envISO, err := ioutil.TempFile("", "agent-env-iso")
	if err != nil {
		return
	}
	envISO.Close()
	output, err := runner.Run("mkisofs", "-o", envISO.Name(), envFile.Name())
	if err != nil {
		out := string(output[:])
		return "", errors.New(fmt.Sprintf("Failed to generate ISO for agent settings: %v\n%s", err, out))
	}
	// Cleanup temp dir but ignore the error. Failure to delete a temp file is not
	// worth worrying about.
	_ = os.RemoveAll(envDir)
	return envISO.Name(), nil
}

// Creates agent env ISO, updates VM metadata, and attaches the ISO to VM
func updateAgentEnv(ctx *cpi.Context, vmID string, env *cpi.AgentEnv) (err error) {
	ctx.Logger.Infof("Creating agent env: %#v", env)
	isoPath, err := createEnvISO(env, ctx.Runner)
	if err != nil {
		return
	}
	defer os.Remove(isoPath)

	// Store env JSON as metadata so it can be picked up by attach/detach disk
	ctx.Logger.Info("Updating metadata for VM")
	err = putAgentEnvMetadata(ctx, vmID, env)
	if err != nil {
		return
	}

	// Detach ISO first, but ignore any task error due to ISO already being detached
	detachTask, err := ctx.Client.VMs.DetachISO(vmID)
	if err != nil && !isTaskError(err) {
		return err
	}
	detachTask, err = ctx.Client.Tasks.Wait(detachTask.ID)
	if err != nil && !isTaskError(err) {
		return err
	}

	ctx.Logger.Infof("Attaching ISO at path: %s", isoPath)
	attachTask, err := ctx.Client.VMs.AttachISO(vmID, isoPath)
	if err != nil {
		return
	}
	ctx.Logger.Infof("Waiting on task: %#v", attachTask)
	attachTask, err = ctx.Client.Tasks.Wait(attachTask.ID)
	return
}
