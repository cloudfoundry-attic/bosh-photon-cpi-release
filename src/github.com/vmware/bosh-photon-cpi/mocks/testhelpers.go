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
	"encoding/json"
)

func ToJson(v interface{}) string {
	res, err := json.Marshal(v)
	if err != nil {
		// Since this method is only for testing, don't return
		// any errors, just panic.
		panic("Error serializing struct into JSON")
	}
	// json.Marshal returns []byte, convert to string
	return string(res[:])
}
