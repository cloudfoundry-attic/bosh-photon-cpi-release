#BOSH Photon CPI written in Go


## Bosh Manifest Example for bosh-photon-cpi

This is a sample of how Photon specific properties are used in a  BOSH deployment manifest:

    ---
	networks:
	- name: default 
	  subnets:
	    reserved:
	    - 192.168.21.129 - 192.168.21.150
	    static:
	    - 192.168.21.151 - 192.168.21.189
	    range: 192.168.21.128/25
	    gateway: 192.168.21.253
	    dns:
	    - 192.168.71.1
	    cloud_networks:
	      name: "cloud_network"

    ...

    properties:
	  ntp:
	  - 192.168.21.10
	  cpi:
	    agent:
	      mbus: nats://nats:nats-password@192.168.21.4:4222
	  photon:
	    target: https://192.168.10.1:8080
	    user: dev
	    password: pwd
	    tenant: dev
	    project: dev
	    description: Bosh on Photon
