#!/usr/bin/env bash

set -ex

WORKSPACE=$(pwd)

ls -l .

# fix bosh-init
mv $WORKSPACE/bosh-init/bosh-init-* $WORKSPACE/bosh-init/bosh-init
chmod +x $WORKSPACE/bosh-init/bosh-init

# setup deployment
cp $WORKSPACE/bosh-deployment-manifest/bosh-micro.yml $WORKSPACE/bosh-deployment

# deploy
$WORKSPACE/bosh-init/bosh-init deploy $WORKSPACE/bosh-deployment/bosh-micro.yml
