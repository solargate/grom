# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Water sport type Packraft
- CI workflow on `master` and pull requests: Go vet/test/build, Flutter analyze/test, and swagger doc drift check

### Changed

- Server CLI moved to Cobra: root starts the server (`--config` / `-c`), `gencerts` is a subcommand; `--help` and `--version` are available

## [0.1.0] - 2026-07-18

First public release.

### Added

- Workout tracking with stats, notes, media, and map previews
- GPS track import (GPX, FIT) and live recording on Android
- Equipment management linked to workouts
- Social feed with local and federated follows (optional ActivityPub)
- Strava bulk import from a data export ZIP
- Embedded web UI served by the Go server and Android APK builds
- Locales: English, Russian, and German
- Unified versioning via `VERSION` (server binary and Flutter `--build-name`)
- Release CI: cross-platform server packages and APK on semver tags

### Changed

- Storage layout refactored behind a `storage.Backend` interface (filesystem driver)
- Workout card UI redesign

### Fixed

- Strava import: convert speed from mph to km/h where applicable

[Unreleased]: https://github.com/solargate/grom/compare/0.1.0...HEAD
[0.1.0]: https://github.com/solargate/grom/releases/tag/0.1.0
