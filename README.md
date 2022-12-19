[![CircleCI](https://circleci.com/gh/giantswarm/crsync.svg?&style=shield)](https://circleci.com/gh/giantswarm/crsync) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/crsync/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/crsync)

# crsync

The `crsync` tool synchronizes images between `quay.io` and configured target container registry.

## Getting Project

Download the latest release:
https://github.com/giantswarm/crsync/releases/latest

Clone the git repository: https://github.com/giantswarm/crsync.git

Download the latest docker image from here:
https://quay.io/repository/giantswarm/crsync


### How to build

Build the standard way.

```
go build github.com/giantswarm/crsync
```

## Installing the Chart

To install the chart locally:

```bash
$ git clone https://github.com/giantswarm/crsync.git
$ cd crsync
$ helm install helm/crsync
```

Provide a custom `values.yaml`:

```bash
$ helm install crsync -f values.yaml
```

Deployment to Tenant Clusters is handled by [app-operator](https://github.com/giantswarm/app-operator).

## Configuration

There are few mandatory configuration options:

```
lastModified: 2h
destinationRegistry:
  name: <container-registry-address> # e.g. docker.io
  credentials:
    user: <container-registry-user>
    password: <base64-encoded-password>
```


## Release Process

* Ensure CHANGELOG.md is up to date.
* Create a new GitHub release with the version e.g. `v0.1.0` and link the
changelog entry.
* This will push a new git tag and trigger a new tarball to be pushed to the
[giantswarm-operations-platform-catalog].

[app-operator]: https://github.com/giantswarm/app-operator
[giantswarm-operations-platform-catalog]: https://github.com/giantswarm/giantswarm-operations-platform-catalog
[giantswarm-operations-platform-catalog-test-catalog]: https://github.com/giantswarm/giantswarm-operations-platform-test-catalog

## License

crsync is under the Apache 2.0 license. See the [LICENSE](LICENSE) file
for details.
