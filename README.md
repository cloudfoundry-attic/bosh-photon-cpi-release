# bosh-photon-cpi-release

A [BOSH](https://github.com/cloudfoundry/bosh) release for `bosh-photon-cpi` written in Go.

### Example environment

`bosh-photon-cpi` release can be deployed with any BOSH Director
just like any other BOSH release.

1. Install Vagrant dependencies

```
vagrant plugin install vagrant-bosh
gem install bosh_cli --no-ri --no-rdoc
```

1. Create a dev release
```
bosh -n create release --force
```

1. Create a new VM with BOSH Director and BOSH Photon CPI releases

```
vagrant up
```

Note: See [deployment manifest](manifests/bosh-micro.yml)
to see how bosh and bosh cpi releases are collocated.

1. Target deployed BOSH Director

```
bosh target localhost:25555
bosh status
```

### Running tests

1. Follow instructions above to install the release to your BOSH director

1. Clone BOSH repository into `$HOME/workspace/bosh` to get BATS source code

1. Download vSphere stemcell #3 to `$HOME/Downloads/`
   from [BOSH Artifacts](http://bosh.io/releases)

1. Run BOSH Acceptance Tests via `spec/run-bats.sh`

