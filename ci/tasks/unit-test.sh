#!/usr/bin/env bash

set -ex

WORKSPACE=$(pwd)

# -------------------------------------------------
# install dependencies
mkdir $WORKSPACE/tools

# get the blobs
cd $WORKSPACE/cpi-release
bosh sync blobs

mkdir $WORKSPACE/tmp
cp -RL blobs/* $WORKSPACE/tmp
cd $WORKSPACE/tmp

# install mkisofs
export BOSH_INSTALL_TARGET=$WORKSPACE/tools
ls -l cpi_mkisofs/
chmod +x ../cpi-release/packages/cpi_mkisofs/packaging
../cpi-release/packages/cpi_mkisofs/packaging
export PATH=$PATH:$WORKSPACE/tools/bin

# install go
export BOSH_INSTALL_TARGET=$WORKSPACE/tools/go
mkdir $BOSH_INSTALL_TARGET

chmod +x ../cpi-release/packages/golang_1.4.2/packaging
../cpi-release/packages/golang_1.4.2/packaging
export GOROOT=$WORKSPACE/tools/go
export PATH=$PATH:$WORKSPACE/tools/go/bin

rm -rf $WORKSPACE/tmp

# -------------------------------------------------
# run bosh-photon-cpi unit tests
cd $WORKSPACE/cpi-release

export GOPATH=$WORKSPACE/cpi-release
go test -v github.com/vmware/bosh-photon-cpi -ginkgo.noColor -ginkgo.slowSpecThreshold=60 -ginkgo.v
