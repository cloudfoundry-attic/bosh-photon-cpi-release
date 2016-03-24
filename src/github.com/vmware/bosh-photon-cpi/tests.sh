# Convenience script to run only unit tests without running integration tests,
# so you don't accidentally run 'go test ./...' and kick off integration tests.
go test -v github.com/vmware/bosh-photon-cpi
