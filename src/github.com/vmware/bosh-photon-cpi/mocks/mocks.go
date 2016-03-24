// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package mocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/vmware/bosh-photon-cpi/cpi"
	. "github.com/vmware/photon-controller-go-sdk/photon"
)

func NewMockServer() (server *httptest.Server) {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, r.URL)
		}))
}

func NewMockTask(operation, state string, id string, steps ...Step) *Task {
	return &Task{Operation: operation, State: state, ID: id, Steps: steps}
}

func CreateResponder(status int, response string) Responder {
	return Responder(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode:    status,
			ProtoMajor:    1,
			ProtoMinor:    0,
			Body:          ioutil.NopCloser(bytes.NewBufferString(response)),
			ContentLength: int64(len(response)),
			Request:       req,
		}

		resp.Header = make(map[string][]string)
		resp.Header.Add("Content-Type", "application/json")

		return resp, nil
	})
}

func GetResponse(data []byte) (res cpi.Response, err error) {
	err = json.Unmarshal(data, &res)
	return
}

func GetEnvMetadata(env *cpi.AgentEnv) (res string) {
	json, err := json.Marshal(env)
	if err != nil {
		panic("Unable to serialize agent env JSON")
	}
	res = string(json[:])
	return
}
