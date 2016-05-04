#!/bin/bash

set -e

CURDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TESTOUT="$CURDIR/testout"
mkdir -p $TESTOUT

if [ -z "$TEST_TARGET" ]; then
  echo "TEST_TARGET needs to be set"
  exit -1
fi

export TEST_STEMCELL_URL=${TEST_STEMCELL_URL:-"https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3184.1"}
export TEST_STEMCELL="$TESTOUT/image"
export PATH=$GOPATH/bin:$PATH

# Only download stemcell if it isn't present, useful for testing
# locally when you don't want to wait for the download
if [ ! -f $TEST_STEMCELL ]; then
	echo "Downloading stemcell from bosh.io"
	wget -O $TESTOUT/vsphere-stemcell.tgz $TEST_STEMCELL_URL
	tar xf $TESTOUT/vsphere-stemcell.tgz -C $TESTOUT
fi
if [ $? -ne 0 ]; then
	echo "Failed to download stemcell from bosh.io"
	exit 1
fi

echo "Running tests"
go test -v github.com/vmware/bosh-photon-cpi/integration
