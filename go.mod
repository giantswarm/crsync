module github.com/giantswarm/crsync

go 1.16

require (
	github.com/containers/image/v5 v5.10.5
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.1.3
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
)

replace (
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
)
