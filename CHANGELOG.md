# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `/api/docs` (no trailing slash) redirects to `/api/docs/` so Swagger UI assets load instead of a blank page

## [0.3.0] - 2026-07-27

### Added

- User and admin documentation under `docs/` (overview with Android screenshots, install, configuration); root README shortened as the entry point
- Structured server logging via `log/slog` with configurable `logging.level` (`debug`/`info`/`warn`/`error`) and `logging.format` (`text`/`json`); HTTP access logs through Gin middleware
- Unexpected API 500s log the underlying error (with request id when available); panic recovery goes through slog
- Auth signals: `login_failed` / `user_registered` / `register_conflict` (passwords never logged)
- Federation publish logs inbox-list failures and track/media read warnings instead of failing silently
- Strava import result reports `media_missing`: photo files referenced in `activities.csv` but absent from the ZIP archive
- Branded app icons and web favicon from the Grom logo (Android, iOS, web)
- Equipment list shows accumulated distance (km) per item; updated automatically when workouts change or after Strava bulk import

### Changed

- **Breaking:** Default autocert ACME cache is `acme-cache` next to the grom binary (no longer `{storage.location}/acme-cache`). Move an existing cache or set `server.tls.autocert.cache_dir` to the old path before upgrading
- Equipment navigation menu icon: inventory (`inventory_2`) instead of sports whistle
- Workout list cards no longer show equipment; detail view shows it below the stats table
- Workout track overlay on interactive maps and static map previews uses orange (`#F45E1E`) instead of blue

### Fixed

- Returning from workout detail keeps the home feed/list scroll position and loaded pages

## [0.2.0] - 2026-07-26

### Added

- `GET /api/v1/workouts/{id}` to fetch a single workout (`?owner=` same as track/media)
- Edit own workouts from the detail menu (metadata and equipment), with ActivityPub `Update` delivery to followers
- Water sport type Packraft
- CI workflow on `master` and pull requests: Go vet/test/build, Flutter analyze/test, and swagger doc drift check
- `storage.driver: bbolt` — metadata in a Bolt DB (JSON), tracks/photos/avatars remain on the filesystem under `storage.location`
- `grom migrate-storage` to copy metadata between `file` and `bbolt` (`--from` / `--to`, optional `--dry-run` / `--verify` / `--force`)

### Changed

- **Breaking:** `GET /api/v1/workouts` returns a cursor page `{items, next_cursor, has_more}` instead of a bare array; query params `limit` (default 20, max 100) and `cursor`
- **Breaking:** `GET /api/v1/server_info` renamed to `GET /api/v1/server-info`
- Server CLI moved to Cobra: root starts the server (`--config` / `-c`), `gencerts` is a subcommand; `--help` and `--version` are available
- Relative TLS cert/key, federation CA, and autocert `cache_dir` paths resolve against the grom binary directory (same as `storage.location` / `temp_dir`)
- Newly allocated workout IDs are unique across all local users on the instance (existing data is left unchanged)

### Fixed

- Workout list enrichment no longer re-scans the workouts directory per item (O(N²) on large libraries)
- Workout create responses now report `has_map_preview` correctly after track attach and after adding photos

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

[Unreleased]: https://github.com/solargate/grom/compare/0.3.0...HEAD
[0.3.0]: https://github.com/solargate/grom/releases/tag/0.3.0
[0.2.0]: https://github.com/solargate/grom/releases/tag/0.2.0
[0.1.0]: https://github.com/solargate/grom/releases/tag/0.1.0
