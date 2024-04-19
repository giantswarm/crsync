# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.4] - 2024-04-19

## [0.9.3] - 2024-01-29

### Fixed

- Move pss values under the global property

## [0.9.2] - 2024-01-09

### Changed

- Configure `gsoci.azurecr.io` as the default container image registry.

## [0.9.1] - 2023-12-05

### Fixed

- Fix Enforced PSS mode check.

## [0.9.0] - 2023-12-04

### Fixed

- Fix Kyverno PolicyException.
- Add seccompProfile.

### Changed

- Move to `app-build-suite` and add team annotation.

## [0.8.0] - 2023-11-28

### Added

- Add Kyverno PolicyException.

## [0.7.0] - 2023-11-10

### Changed

- Add a switch for PSP CR installation.
- Add 60 seconds timeout to metrics endpoint to solve reported issue by `golangcli-lint`: [G114](https://github.com/golangci/golangci-lint/blob/1926748b44fb0dbca8c320bf2145190367d7fedc/.golangci.reference.yml#L774)

## [0.6.1] - 2022-02-24

### Fixed

- Handle pagination when listing repositories using the Quay REST API.

## [0.6.0] - 2021-05-20

### Added

- Limit number of retag jobs executed in parallel to 4 to prevent docker from choking.
- Print overall progress every minute.

### Fixed

- Fix error when destination (Docker Hub) repository doesn't have any tags yet.

## [0.5.11] - 2021-02-02

### Added

- Print errors pretty.

### Fixed

- List Docker Hub tags with Registry V2 API to avoid hitting limits.

## [0.5.10] - 2021-01-29

### Fixed

- Handle error responses on listing tags from Docker Hub.
- Calculate next endpoint based on number of tags in repository.

## [0.5.9] - 2021-01-28

### Changed

- Increase memory limit from 100mb to 500mb.

## [0.5.8] - 2020-07-29

### Changed

- Decrease burst for parallel retagging.
- Increase pull/push rate limiter configuration.

### Fixed

- Fix metrics for number of tags in destination registry repository.

## [0.5.7] - 2020-07-23

### Fixed

- Remove hostPort.

## [0.5.6] - 2020-07-23

### Fixed

- Shorten port names to fit <15 chars convention.

## [0.5.5] - 2020-07-23

## [0.5.4] - 2020-07-23


### Fixed

- Fix listing tags in `quay.io` registry.
- Fix deployment's metrics port definition.

## [0.5.3] - 2020-07-23

### Fixed

- Fix network policy and service to be visible to Prometheus as a target.

## [0.5.2] - 2020-07-22

### Added

- Add `quay` container registry authorization.
- Add synchronization for private repositories.

## [0.5.1] - 2020-07-22

### Fixed

- Measure tag count in case of **no** errors.

## [0.5.0] - 2020-07-21

### Added

- Expose Prometheus metrics for `sync` command.

### Changed

- Stay logged in between jobs for at least 24h.

### Fixed

- Fix gauge to be updated only on successful tag counts.

## [0.4.1] - 2020-07-17

### Fixed

- Fix the issue where too many tags lead to a deadlock by starting both
  processors (tag listing & retagging) in parallel.
- Make helm resource names unique per release.
- Reduce number of concurrent push/pull operations to avoid docker client
  kills.

## [0.4.0] - 2020-07-17

### Added

- Add `--loop` flag to `sync` command allowing to run it continuously.
- Run operations against docker registry in parallel in `sync` command.

### Changed

- Run as a continuous service in a Deployment instead of a CronJob.
- Move synchronization logic into `sync` sub-command.
- Synchronize all tags instead of just releases.

### Fixed

- Fix tags listing in azure container registry.
- Fix tags listing in dockerhub container registry.

## [0.3.0] - 2020-07-09

### Changed

- Use authenticated API calls to get list of existing tags in destination repository.

## [0.2.0] - 2020-07-03

### Changed

- Replace `v1` registry endpoint call with `docker pull` command to check if requested image tag exists.

## [0.1.0] - 2020-07-02

### Added

- Add initial code.
- Add first version of the helm chart.
- Add release automation.

[Unreleased]: https://github.com/giantswarm/crsync/compare/v0.9.4...HEAD
[0.9.4]: https://github.com/giantswarm/crsync/compare/v0.9.3...v0.9.4
[0.9.3]: https://github.com/giantswarm/crsync/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/giantswarm/crsync/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/giantswarm/crsync/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/giantswarm/crsync/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/crsync/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/giantswarm/crsync/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/giantswarm/crsync/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/giantswarm/crsync/compare/v0.5.11...v0.6.0
[0.5.11]: https://github.com/giantswarm/crsync/compare/v0.5.10...v0.5.11
[0.5.10]: https://github.com/giantswarm/crsync/compare/v0.5.9...v0.5.10
[0.5.9]: https://github.com/giantswarm/crsync/compare/v0.5.8...v0.5.9
[0.5.8]: https://github.com/giantswarm/crsync/compare/v0.5.7...v0.5.8
[0.5.7]: https://github.com/giantswarm/crsync/compare/v0.5.6...v0.5.7
[0.5.6]: https://github.com/giantswarm/crsync/compare/v0.5.5...v0.5.6
[0.5.5]: https://github.com/giantswarm/crsync/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/giantswarm/crsync/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/giantswarm/crsync/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/giantswarm/crsync/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/giantswarm/crsync/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/crsync/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/giantswarm/crsync/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/giantswarm/crsync/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/crsync/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/crsync/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/crsync/releases/tag/v0.1.0
