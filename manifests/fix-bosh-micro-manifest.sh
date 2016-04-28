#!/bin/bash

set -ex

# Check if all variables in bosh-micro manifest have been set
if [ -z "$BOSH_CPI_RELEASE_FILE_NAME" ]; then
  echo "BOSH_CPI_RELEASE_FILE_NAME needs to be set"
  exit -1
fi
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
if [ -z "$PHOTON_TARGET" ]; then
  echo "PHOTON_TARGET needs to be set"
  exit -1
fi
if [ -z "$PHOTON_PROJECT_ID" ]; then
  echo "PHOTON_PROJECT_ID needs to be set"
  exit -1
fi

# replace all env variables in bosh-micro.yml and generate new bosh-micro-fixed.yml file
cd $(dirname $0)
eval "echo \"$(< ./bosh-micro.yml)\"" > ./bosh-micro-fixed.yml
