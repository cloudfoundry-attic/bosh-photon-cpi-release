#!/bin/bash

set -e

director_target=10.146.38.111
director_user=admin
director_password=admin

bosh -n target $director_target
bosh -n login $director_user $director_password

director_uuid=$(bosh -n status --uuid)

spec_path=$PWD/bat.spec

# Create bat.spec used by BATS to generate BOSH manifest
cat > $spec_path << EOF
---
cpi: photon
manifest_template_path: $(echo `pwd`/photon.yml.erb)
properties:
  uuid:  $BOSH_UUID
  pool_size: 1
  instances: 1
  second_static_ip: 10.146.39.12
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: "3184.1"
  mbus: nats://nats:nats-password@10.146.38.111:4222
  networks:
  - name: static
    type: manual
    cidr: 10.146.39.0/24
    static_ip: 10.146.39.9
    static:
      - 10.146.39.1 - 10.146.39.15
    reserved:
      - 10.146.39.100 - 10.146.39.200
    gateway: 10.146.39.253
    dns: [10.132.71.1]
    vlan: "VM VLAN"
  - name: second
    type: manual
    cidr: 10.146.39.0/24
    static_ip: 10.146.39.19
    static:
      - 10.146.39.1 - 10.146.39.25
    reserved:
      - 10.146.39.100 - 10.146.39.200
    gateway: 10.146.39.253
    dns: [10.132.71.1]
    vlan: "VM VLAN"
EOF

# Set up env vars used by BATS
export BAT_DEPLOYMENT_SPEC=$spec_path
export BAT_STEMCELL=$HOME/src/bosh-stemcell-3184.1-vsphere-esxi-ubuntu-trusty-go_agent.tgz
export BAT_DIRECTOR=$director_target
export BAT_DNS_HOST=$director_target
export BAT_VCAP_PASSWORD=c1oudc0w
export BAT_INFRASTRUCTURE=photon
export BAT_NETWORKING=manual
export BAT_SECOND_STATIC_IP=10.146.39.12

cd ../src/bosh/bat

bundle install
bundle exec rspec spec --tag ~multiple_manual_networks
