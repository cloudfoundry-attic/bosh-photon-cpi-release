#!/bin/bash

set -e

# Check if all variables in manifest have been set
if [ -z "$DIRECTOR_IP" ]; then
  echo "DIRECTOR_IP needs to be set"
  exit -1
fi
if [ -z "$NETWORK_RANGE" ]; then
  echo "NETWORK_RANGE needs to be set"
  exit -1
fi
if [ -z "$NETWORK_GATEWAY" ]; then
  echo "NETWORK_GATEWAY needs to be set"
  exit -1
fi
if [ -z "$NETWORK_ID" ]; then
  echo "NETWORK_ID needs to be set"
  exit -1
fi
if [ -z "$BAT_STATIC_IP" ]; then
  echo "BAT_STATIC_IP needs to be set"
  exit -1
fi
if [ -z "$BAT_SECOND_STATIC_IP" ]; then
  echo "BAT_SECOND_STATIC_IP needs to be set"
  exit -1
fi
if [ -z "$BAT_IP_RANGE" ]; then
  echo "BAT_IP_RANGE needs to be set"
  exit -1
fi
if [ -z "$BAT_IP_RESERVED" ]; then
  echo "BAT_IP_RESERVED needs to be set"
  exit -1
fi
if [ -z "$STEMCELL_PATH" ]; then
  echo "STEMCELL_PATH needs to be set"
  exit -1
fi
if [ -z "$BAT_NETWORK_RANGE_2" ]; then
  echo "BAT_NETWORK_RANGE_2 needs to be set"
  exit -1
fi
if [ -z "$BAT_STATIC_IP_2" ]; then
  echo "BAT_STATIC_IP_2 needs to be set"
  exit -1
fi
if [ -z "$BAT_IP_RANGE_2" ]; then
  echo "BAT_IP_RANGE_2 needs to be set"
  exit -1
fi
if [ -z "$BAT_IP_RESERVED_2" ]; then
  echo "BAT_IP_RESERVED_2 needs to be set"
  exit -1
fi
if [ -z "$BAT_GATEWAY_2" ]; then
  echo "BAT_GATEWAY_2 needs to be set"
  exit -1
fi
if [ -z "$BAT_NETWORK_ID_2" ]; then
  echo "BAT_NETWORK_ID_2 needs to be set"
  exit -1
fi

director_target=$DIRECTOR_IP
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
  uuid:  $director_uuid
  pool_size: 1
  instances: 1
  second_static_ip: $BAT_SECOND_STATIC_IP
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: "3184.1"
  mbus: nats://nats:nats-password@$DIRECTOR_IP:4222
  networks:
  - name: static
    type: manual
    cidr: $NETWORK_RANGE
    static_ip: $BAT_STATIC_IP
    static: [$BAT_IP_RANGE]
    reserved: [$BAT_IP_RESERVED]
    gateway: $NETWORK_GATEWAY
    dns: [10.132.71.1]
    vlan: $NETWORK_ID
  - name: second
    type: manual
    cidr: $BAT_NETWORK_RANGE_2
    static_ip: $BAT_STATIC_IP_2
    static: [$BAT_IP_RANGE_2]
    reserved: [$BAT_IP_RESERVED_2]
    gateway: $BAT_GATEWAY_2
    dns: [10.132.71.1]
    vlan: $BAT_NETWORK_ID_2
EOF

# Set up env vars used by BATS
export BAT_DEPLOYMENT_SPEC=$spec_path
export BAT_STEMCELL=$STEMCELL_PATH
export BAT_DIRECTOR=$director_target
export BAT_DNS_HOST=$director_target
export BAT_VCAP_PASSWORD=c1oudc0w
export BAT_INFRASTRUCTURE=photon
export BAT_NETWORKING=manual

cd ../src/bosh/bat

bundle install
bundle exec rspec spec --tag ~multiple_manual_networks
