# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Make helm resource names unique per release.

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

[Unreleased]: https://github.com/giantswarm/crsync/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/giantswarm/crsync/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/crsync/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/crsync/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/crsync/releases/tag/v0.1.0
