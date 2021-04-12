module github.com/giantswarm/crsync

go 1.14

require (
	github.com/containers/image/v5 v5.11.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.3
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
)

replace github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
