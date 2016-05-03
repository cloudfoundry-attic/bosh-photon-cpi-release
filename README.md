# bosh-photon-cpi-release

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* IRC: [`#bosh` on freenode](https://webchat.freenode.net/?channels=bosh)
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: to be released
* Roadmap: [Pivotal Tracker] to be released

A [BOSH](https://github.com/cloudfoundry/bosh) release for `bosh-photon-cpi` written in Go.



### Example setup

#### Install the BOSH CLI

Install the [BOSH CLI](http://bosh.io/docs/bosh-cli.html) tool.

#### Install the bosh-init CLI

Install the [bosh-init](https://bosh.io/docs/install-bosh-init.html) tool.

#### Create a new development BOSH release

1. Make local changes to the release

2. Create a dev release, (--force is required if there are local changes not committed to git)

    ```
    bosh create release --force --with-tarball
    ```

#### Create a deployment manifest

Create a `photon-bosh.yml` deployment manifest file.
See [deployment manifest](manifests/bosh-micro.yml) for an example manifest. Update it with your properties.
See [PHOTON_CPI.md](PHOTON_CPI.md) for cloud properties description.

#### Deploy

Using the previously created deployment manifest, now we can deploy it:

```
$ bosh-init deploy photon-bosh.yml
```

Then target your deployed BOSH director. Your default username is `admin` and password is `admin`.

```
$ bosh target <YOUR BOSH IP ADDRESS>
$ bosh status
```



### Running unit tests

1. Set your GOPATH to "bosh-photon-cpi-release" folder

2. Run BOSH Unit Tests via `src/github.com/vmware/bosh-photon-cpi/tests.sh`



### Running BATS

1. Follow instructions above to install the release to your BOSH director

2. Clone BOSH repository into `$HOME/workspace/bosh` to get BATS source code

3. Download vSphere ubuntu stemcell 3184.1 version to `$HOME/Downloads/`
   from [BOSH Artifacts](http://bosh.io/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent)

4. Set all variables required in `spec/run-bats.sh`. Run BOSH Acceptance Tests via `spec/run-bats.sh`

