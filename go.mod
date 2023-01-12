module github.com/giantswarm/crsync

go 1.16

// https://github.com/containerd/containerd/issues/5781
exclude k8s.io/kubernetes v1.13.0

require (
	github.com/containers/image/v5 v5.23.1
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v1.0.0
	github.com/prometheus/client_golang v1.14.0
	github.com/spf13/cobra v1.6.1
	golang.org/x/sync v0.1.0
	golang.org/x/time v0.3.0
)

replace (
	github.com/aws/aws-sdk-go v1.27.0 => github.com/aws/aws-sdk-go v1.44.53
	github.com/containerd/containerd v1.3.2 => github.com/containerd/containerd v1.6.6
	github.com/containerd/containerd v1.6.1 => github.com/containerd/containerd v1.6.6
	github.com/containers/storage v1.24.8 => github.com/containers/storage v1.40.3
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/consul/api v1.3.0 => github.com/hashicorp/consul/api v1.13.0
	github.com/hashicorp/consul/sdk v0.3.0 => github.com/hashicorp/consul/sdk v0.9.0
	github.com/miekg/dns v1.0.14 => github.com/miekg/dns v1.1.50
	github.com/nats-io/nats-server/v2 v2.1.2 => github.com/nats-io/nats-server/v2 v2.8.4
	github.com/opencontainers/runc v1.0.0-rc91 => github.com/opencontainers/runc v1.1.3
	// Resolves sonatype-2019-0890
	github.com/pkg/sftp v1.10.1 => github.com/pkg/sftp v1.13.5
	github.com/prometheus/client_golang v1.10.0 => github.com/prometheus/client_golang v1.12.2
)
