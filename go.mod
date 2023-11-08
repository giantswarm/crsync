module github.com/giantswarm/crsync

go 1.19

// https://github.com/containerd/containerd/issues/5781
exclude k8s.io/kubernetes v1.13.0

require (
	github.com/containers/image/v5 v5.28.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v1.0.0
	github.com/prometheus/client_golang v1.17.0
	github.com/spf13/cobra v1.8.0
	golang.org/x/sync v0.5.0
	golang.org/x/time v0.4.0
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containers/libtrust v0.0.0-20230121012942-c1716e8a8d01 // indirect
	github.com/containers/ocicrypt v1.1.8 // indirect
	github.com/containers/storage v1.50.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.6+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/opencontainers/runc v1.1.9 // indirect
	github.com/opencontainers/runtime-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/vbatts/tar-split v0.11.5 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/aws/aws-sdk-go v1.27.0 => github.com/aws/aws-sdk-go v1.44.53
	github.com/cloudflare/circl => github.com/cloudflare/circl v1.3.6
	github.com/containerd/containerd v1.3.2 => github.com/containerd/containerd v1.6.6
	github.com/containerd/containerd v1.6.1 => github.com/containerd/containerd v1.6.6
	github.com/containers/storage v1.24.8 => github.com/containers/storage v1.40.3
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v5 v5.1.0
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/consul/api v1.3.0 => github.com/hashicorp/consul/api v1.13.0
	github.com/hashicorp/consul/sdk v0.3.0 => github.com/hashicorp/consul/sdk v0.9.0
	github.com/miekg/dns v1.0.14 => github.com/miekg/dns v1.1.50
	github.com/nats-io/nats-server/v2 v2.1.2 => github.com/nats-io/nats-server/v2 v2.8.4
	github.com/opencontainers/runc v1.0.0-rc91 => github.com/opencontainers/runc v1.1.3
	// Resolves sonatype-2019-0890
	github.com/pkg/sftp v1.10.1 => github.com/pkg/sftp v1.13.5
	github.com/prometheus/client_golang v1.10.0 => github.com/prometheus/client_golang v1.12.2
	golang.org/x/net => golang.org/x/net v0.17.0
	google.golang.org/grpc => google.golang.org/grpc v1.59.0
)
