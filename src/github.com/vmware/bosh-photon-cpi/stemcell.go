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
	"github.com/vmware/photon-controller-go-sdk/photon"
	"os"
	"path/filepath"
)

func CreateStemcell(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	imagePath, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where image_path should be")
	}

	ctx.Logger.Infof("CreateStemcell with imagePath: '%s'", imagePath)

	ctx.Logger.Info("Reading stemcell from disk")
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return
	}
	defer imageFile.Close()

	defaultReplication := "EAGER"
	options := &photon.ImageCreateOptions{
		ReplicationType: defaultReplication,
	}

	ctx.Logger.Info("Beginning stemcell upload")
	task, err := ctx.Client.Images.Create(imageFile, filepath.Base(imagePath), options)
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

func DeleteStemcell(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("Expected at least 1 argument")
	}
	stemcellCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where stemcell_cid should be")
	}

	ctx.Logger.Infof("DeleteStemcell with stemcell_cid: '%s'", stemcellCID)

	ctx.Logger.Info("Beginning stemcell deletion")
	task, err := ctx.Client.Images.Delete(stemcellCID)
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
