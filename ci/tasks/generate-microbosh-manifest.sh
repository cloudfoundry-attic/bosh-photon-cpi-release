#!/usr/bin/env bash
set -ex

WORKSPACE=$(pwd)

ls -l .

if [ -z "PHOTON_CLI_URL" ]; then
  echo "PHOTON_CLI_URL needs to be set"
  exit -1
fi

if [ -z "$PHOTON_TARGET" ]; then
  echo "PHOTON_TARGET needs to be set"
  exit -1
fi

if [ -z "PHOTON_TENANT" ]; then
  echo "PHOTON_TENANT needs to be set"
  exit -1
fi

if [ -z "PHOTON_PROJECT" ]; then
  echo "PHOTON_PROJECT needs to be set"
  exit -1
fi

if [ -z "PHOTON_NETWORK" ]; then
  echo "PHOTON_NETWORK needs to be set"
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

if [ -z "$NETWORK_DNS" ]; then
  echo "NETWORK_DNS needs to be set"
  exit -1
fi


# get the cli
wget -q -N ${PHOTON_CLI_URL} -o $WORKSPACE/photon
chmod +x $WORKSPACE/photon

# configure the CLI
$WORKSPACE/photon -n target set -c ${PHOTON_TARGET}

# determine project id
export PHOTON_PROJECT_ID=$($WORKSPACE/photon -n project list -t ${PHOTON_TENANT} | grep ${PHOTON_PROJECT} | awk  -F " " '{print$1}')

# determine network id
export NETWORK_ID=$($WORKSPACE/photon -n network list -n ${PHOTON_NETWORK} | awk -F " " '{print$1}')

# determine bosh release file path
export BOSH_RELEASE_FILE_PATH=$WORKSPACE/bosh-release/release.tgz

# determine cpi release file path
export BOSH_CPI_RELEASE_FILE_PATH=$(ls -l $WORKSPACE/cpi-release-candidate/bosh-photon-cpi-*.tgz | awk -F " " '{print$9}')

# determine stemcell file path
export BOSH_STEMCELL_FILE_PATH=$WORKSPACE/stemcell/stemcell.tgz

# fix up bosh deployment manifest
cd $(dirname $0)/../manifests
eval "echo \"$(< ./bosh-micro-template.yml)\"" > ./bosh-micro.yml
