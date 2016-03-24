#!/bin/bash

TESTOUT="./testout"
mkdir -p $TESTOUT
export TEST_TARGET=${TEST_TARGET:-"http://10.146.38.100:8080"}
export TEST_STEMCELL_URL=${TEST_STEMCELL_URL:-"https://ci.ec.eng.vmware.com/job/bosh-stemcell/lastSuccessfulBuild/artifact/tmp/bosh-stemcell-0000-vsphere-esxi-ubuntu-trusty-go_agent.tgz"}
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
go test -v github.com/vmware/bosh-photon-cpi/inttests
