#!/bin/bash

set -e

CURDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TESTOUT="$CURDIR/testout"
mkdir -p $TESTOUT

if [ -z "$TEST_TARGET" ]; then
  echo "TEST_TARGET needs to be set"
  exit -1
fi

export TEST_STEMCELL_URL=${TEST_STEMCELL_URL:-"http://artifactory.ec.eng.vmware.com:3000/bosh-stemcell-3184.1-vsphere-esxi-ubuntu-trusty-go_agent.tgz"}
export TEST_STEMCELL="$TESTOUT/image"
export PATH=$GOPATH/bin:$PATH

# Only download stemcell if it isn't present, useful for testing
# locally when you don't want to wait for the download
if [ ! -f $TEST_STEMCELL ]; then
	echo "Downloading stemcell from Jenkins"
	wget -O $TESTOUT/vsphere-stemcell.tgz $TEST_STEMCELL_URL
	tar xf $TESTOUT/vsphere-stemcell.tgz -C $TESTOUT
fi
if [ $? -ne 0 ]; then
	echo "Failed to download stemcell from Jenkins"
	exit 1
fi

echo "Running tests"
go test -v github.com/vmware/bosh-photon-cpi/integration
