module github.com/giantswarm/crsync

go 1.16

require (
	github.com/containers/image/v5 v5.16.1
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	golang.org/x/sync$2036812b2e83c
	golang.org/x/time$21f47c861a9ac
)

replace (
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
)
