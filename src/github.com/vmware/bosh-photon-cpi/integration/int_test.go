// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package inttests

import (
	"encoding/json"
	"fmt"
	ec "github.com/vmware/photon-controller-go-sdk/photon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
)

var (
	configPath string
	tenant     string
	project    string
	vmFlavor   string
	diskFlavor string
	client     *ec.Client
)

func ECSetup() (err error) {
	tenantSpec := &ec.TenantCreateSpec{Name: uniqueName("bosh-cpi-int-tenant")}
	tenantTask, err := client.Tenants.Create(tenantSpec)
	if err != nil {
		return
	}
	// Make sure to print tenant, project, etc info to aid with debugging
	fmt.Printf("Tenant ID: %s\nTenant Name: %s\n", tenantTask.Entity.ID, tenantSpec.Name)

	resSpec := &ec.ResourceTicketCreateSpec{
		Name: uniqueName("bosh-cpi-int-res"),
		Limits: []ec.QuotaLineItem{
			ec.QuotaLineItem{Unit: "GB", Value: 16, Key: "vm.memory"},
			ec.QuotaLineItem{Unit: "COUNT", Value: 10, Key: "vm"},
		},
	}
	_, err = client.Tenants.CreateResourceTicket(tenantTask.Entity.ID, resSpec)
	if err != nil {
		return
	}

	projSpec := &ec.ProjectCreateSpec{
		Name: uniqueName("bosh-cpi-int-project"),
		ResourceTicket: ec.ResourceTicketReservation{
			resSpec.Name,
			[]ec.QuotaLineItem{
				ec.QuotaLineItem{"GB", 4, "vm.memory"},
				ec.QuotaLineItem{"COUNT", 1, "vm"},
			},
		},
	}

	projTask, err := client.Tenants.CreateProject(tenantTask.Entity.ID, projSpec)
	if err != nil {
		return
	}
	fmt.Printf("Project ID: %s\nProject Name: %s\n", projTask.Entity.ID, projSpec.Name)

	diskFlavorSpec := &ec.FlavorCreateSpec{
		Name: uniqueName("bosh-cpi-int-ephdisk"),
		Kind: "ephemeral-disk",
		Cost: []ec.QuotaLineItem{ec.QuotaLineItem{"COUNT", 1, "ephemeral-disk.cost"}},
	}
	_, err = client.Flavors.Create(diskFlavorSpec)
	if err != nil {
		return
	}

	vmFlavorSpec := &ec.FlavorCreateSpec{
		Name: uniqueName("bosh-cpi-int-vm"),
		Kind: "vm",
		Cost: []ec.QuotaLineItem{
			ec.QuotaLineItem{"GB", 2, "vm.memory"},
			ec.QuotaLineItem{"COUNT", 2, "vm.cpu"},
		},
	}
	_, err = client.Flavors.Create(vmFlavorSpec)
	if err != nil {
		return
	}

	tenant = tenantTask.Entity.ID
	project = projTask.Entity.ID
	vmFlavor = vmFlavorSpec.Name
	diskFlavor = diskFlavorSpec.Name
	return
}

var _ = BeforeSuite(func() {
	// Create JSON config file used by CPI. Agent config is unused for this test suite.
	configJson := `{
	"photon": {
		"target": "%s",
		"project": "%s",
		"tenant": "%s"
	},
	"agent": {
		"mbus": "nats://user:pass@127.0.0.1",
		"ntp": ["127.0.0.1"]
	}
}`
	target := os.Getenv("TEST_TARGET")
	client = ec.NewClient(target, nil)
	fmt.Printf("Target: %s\n", target)

	err := ECSetup()
	if err != nil {
		panic(err)
	}

	configJson = fmt.Sprintf(configJson, target, project, tenant)
	configFile, err := ioutil.TempFile("", "bosh-cpi-config")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	_, err = configFile.WriteString(configJson)
	if err != nil {
		panic(err)
	}
	configPath = configFile.Name()
})

var _ = AfterSuite(func() {
	os.Remove(configPath)

	task, err := client.Projects.Delete(project)
	if err != nil {
		fmt.Println("Unable to delete test project")
		fmt.Println(err)
	}
	task, err = client.Tasks.Wait(task.ID)
	if err != nil {
		fmt.Println("Unable to delete test project")
		fmt.Println(err)
	}

	task, err = client.Tenants.Delete(tenant)
	if err != nil {
		fmt.Println("Unable to delete test tenant")
		fmt.Println(err)
	}
	task, err = client.Tasks.Wait(task.ID)
	if err != nil {
		fmt.Println("Unable to delete test tenant")
		fmt.Println(err)
	}
})

var _ = Describe("Bosh CPI", func() {
	It("creates a stemcell successfully", func() {
		stemcell := os.Getenv("TEST_STEMCELL")
		request := `{ "method": "create_stemcell", "arguments": [ "%s", { "vm_flavor": "%s", "disk_flavor": "%s" } ] }`
		request = fmt.Sprintf(request, stemcell, vmFlavor, diskFlavor)

		cmd := exec.Command("bosh-photon-cpi", "-configPath="+configPath)

		stdin, err := cmd.StdinPipe()
		Expect(err).NotTo(HaveOccurred())

		stdout, err := cmd.StdoutPipe()
		Expect(err).NotTo(HaveOccurred())

		// Starts command but does not block
		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		_, err = io.WriteString(stdin, request)
		Expect(err).NotTo(HaveOccurred())
		err = stdin.Close()
		Expect(err).NotTo(HaveOccurred())

		resBytes, err := ioutil.ReadAll(stdout)
		fmt.Println("response is: " + string(resBytes[:]))
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		response := &Response{}
		err = json.Unmarshal(resBytes, response)

		Expect(response.Error).To(BeNil())
		fmt.Println(response)
	})
})

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, rand.Uint32())
}

type BoshErrorType string

const (
	CloudError          BoshErrorType = "Bosh::Clouds::CloudError"
	CpiError            BoshErrorType = "Bosh::Clouds::CpiError"
	NotImplementedError BoshErrorType = "Bosh::Clouds::NotImplemented"
	NotSupportedError   BoshErrorType = "Bosh::Clouds::NotSupported"
)

type Response struct {
	Result interface{}    `json:"result,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
	Log    string         `json:"log,omitempty"`
}

type ResponseError struct {
	Type     BoshErrorType `json:"type"`
	Message  string        `json:"message"`
	CanRetry bool          `json:"ok_to_retry"`
}
