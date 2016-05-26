#!/usr/bin/env bash

set -ex

WORKSPACE=$(pwd)

ls -l .

# fix bosh-init
mv $WORKSPACE/bosh-init/bosh-init-* $WORKSPACE/bosh-init/bosh-init
chmod +x $WORKSPACE/bosh-init/bosh-init

# delete previous deployment
$WORKSPACE/bosh-init/bosh-init delete $WORKSPACE/bosh-deployment-manifest/bosh-micro.yml || echo "No previous deployment found."
