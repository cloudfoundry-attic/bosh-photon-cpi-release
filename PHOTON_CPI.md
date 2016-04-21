---
Photon Platform CPI
---

This topic describes cloud properties for different resources created by the Photon Platform CPI.

## <a id='azs'></a> AZs

Currently the CPI does not support any cloud properties for AZs. Any AZ specification for a job will not have any effect.

Example:

```yaml
azs:
- name: z1
```

---
## <a id='networks'></a> Networks

The CPI only supports manual network subnets.

Schema for `cloud_properties` section used by manual network subnet:

* **network_id** [String, required]: Network resource id in which the instance will be created. Example: `5233c704-15eb-4441-880d-b29085105b35`.

Example of manual network:

```yaml
networks:
- name: default
  type: manual
  subnets:
  - range: 10.10.0.0/24
    gateway: 10.10.0.1
    cloud_properties:
      network_id: 5233c704-15eb-4441-880d-b29085105b35
```

---
## <a id='resource-pools'></a> Resource Pools / VM Types

Schema for `cloud_properties` section:

* **vm_flavor** [String, required]: Name of the `vm` flavor to use to create the instance. This will determine the amount of Memory and number of CPUs for the instance. Example: `core-200`.
* **disk_flavor** [String, required]: Name of the `ephemeral-disk` flavor to use to create all ephemeral disks for the instance. Example: `core-200`.
* **vm_attached_disk_size_gb** [Integer, optional]: Size in GB of the ephemeral-disk attached to the instance. (This is not the boot disk). Default: 16GB. Example: `2`.

Example of an `core-200` instance:

```yaml
resource_pools:
- name: default
  network: default
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: latest
  cloud_properties:
    vm_flavor: core-200
    disk_flavor: core-200
    vm_attached_disk_size_gb: 2
```

---
## <a id='disk-pools'></a> Disk Pools / Disk Types

Schema for `cloud_properties` section:

* **disk_flavor** [String, required]: Name of the `ephemeral-disk` flavor to use to create all ephemeral disks for the instance. Example: `core-200`.

Example of 10GB disk:

```yaml
disk_pools:
- name: default
  disk_size: 10_240
  cloud_properties:
    disk_flavor: core-200
```

---
## <a id='global'></a> Global Configuration

The CPI can only talk a single Photon platform instance and manage VMs within a single **project**.

Schema:

* **target** [String, required]: The fully qualified URL for to the API endpoint of the Photon platform instance. Example: `http://10.0.0.1:8080`
* **ignore\_cert** [Boolean, optional]: Flag that specifies is SSL cert verification should be disabled. Example: `false`
* **project** [String, required]: Id of the **project** resource that VMs should be created under. Example: `2a2d1de6-a28d-4300-b736-7e5bbb6f666f`
* **user** [String, optional]: Username for the API access that has permissions to create VMs within the specified
password. Example: `usr`.
* **password** [String, optional]: Password for the API access. Example: `password`


Example with hard-coded credentials:

```yaml
properties:
  photon:
    target: http://10.0.0.1:8080
    ignore_cert: false
    project: 2a2d1de6-a28d-4300-b736-7e5bbb6f666f
    user: usr
    password: password
```

Example without credentials:

```yaml
properties:
  photon:
    target: http://10.0.0.1:8080
    project: 2a2d1de6-a28d-4300-b736-7e5bbb6f666f
```

---
## <a id='cloud-config'></a> Example Cloud Config

```yaml
azs:
- name: z1

vm_types:
- name: default
  cloud_properties:
    vm_flavor: core-200
    disk_flavor: core-200
- name: large
  cloud_properties:
    vm_flavor: core-200
    disk_flavor: core-200

disk_types:
- name: default
  disk_size: 3000
  cloud_properties: {disk_flavor: core-200}
- name: large
  disk_size: 50_000
  cloud_properties: {disk_flavor: core-200}

networks:
- name: default
  type: manual
  subnets:
  - range: 10.10.0.0/24
    gateway: 10.10.0.1
    static: [10.10.0.62]
    dns: [10.10.0.2]
    cloud_properties: {network_id: 5233c704-15eb-4441-880d-b29085105b35}

compilation:
  workers: 5
  reuse_compilation_vms: true
  vm_type: large
  network: default
```
