#!/usr/bin/env bash

set -ex

version=`cat release-version/number`

cd cpi-release

echo "using bosh CLI version..."
bosh version

cpi_release_name="bosh-photon-cpi"

echo "building CPI release..."
bosh create release --name $cpi_release_name --version $version --with-tarball

mv dev_releases/$cpi_release_name/$cpi_release_name-$version.* ../cpi-release-candidate/
