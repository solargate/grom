# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Account deletion from the profile menu: password confirmation, permanent wipe of user data (file and bbolt), ActivityPub `Delete` of the actor (best-effort), and incoming `Delete` Person handling in the federation inbox
- Profile screen: overflow menu (edit profile, delete account)
- Documentation site on GitHub Pages (MkDocs Material): full `docs/` published at `https://solargate.github.io/grom/`, including Privacy Policy at `/privacy/`

### Changed

- `grom migrate-storage` copies personal access tokens between `file` and `bbolt`; password-reset tokens remain excluded
- `grom migrate-storage --verify` also compares profiles, speed/heart-rate chart counts, and PAT counts; legacy plain-text Like activity ids for local workouts are reconstructed when `federation.domain` is set
- Privacy policy source moved to `docs/privacy.md`; root `PRIVACY.md` is a stub with links to the docs and Pages URL

## [0.8.0] - 2026-08-13

### Added

- Personal access tokens (PAT): scoped API access for workouts and equipment via `grom_pat_…` tokens; `GET/POST /api/v1/auth/pat`, `DELETE /api/v1/auth/pat/{id}`; storage in `personal_access_tokens.yaml` / bbolt `personal_access_tokens`; Integration screen tab **Grom** for token management (tab **External services** for Strava and Health Sync)
- Tests: chart Y-axis helper, login/forgot/reset password and create-workout defaults (Flutter); migrate reset-token exclusion, file profile store, mailer SMTP validation, federation delivery and social search edge cases (Go)

### Changed

- Heart rate and speed chart Y axes start at `max(0, series min − 5)` instead of 0, so the plotted range is easier to read
- Gin debug mode (`[GIN-debug]`) only when `logging.level` is `debug`; otherwise release mode

### Fixed

- Mobile add-workout form: stop clipping the floating “Workout name” label at the top of the Manual tab

## [0.7.3] - 2026-08-09

### Fixed

- Android Health Sync: stop calling web-only `canAccessScopes` after Google Sign-In (it threw `UnimplementedError` and showed a generic sign-in failure); surface technical error detail in the snackbar when present

## [0.7.2] - 2026-08-09

### Fixed

- Android Health Sync: request Google Drive readonly scope explicitly and reconnect once after `invalid_token`, so Play/debug builds do not sync with a bare sign-in token

## [0.7.1] - 2026-08-09

### Changed

- Swagger: shared `ErrorResponse` example is now a neutral `"bad request"`; `@Failure` responses include short per-endpoint descriptions
- Swagger UI Try it out uses the current page origin (no hardcoded `localhost:8080`), so docs work via IP and domain as well as localhost
- Changes in CI (publishing in Google Play)

## [0.7.0] - 2026-08-08

### Added

- Optional ALTCHA captcha (`auth.captcha`, default off): proof-of-work on register, login, and password forgot; `GET /api/v1/captcha/challenge`, `captcha_enabled` on `/server-info`, Flutter checkbox widget
- Password reset via email: `POST /api/v1/auth/password/forgot` and `/reset`, opaque tokens (`reset_tokens.yaml` / bbolt `reset_tokens`), `mailer` config (log/smtp via go-mail, no local MTA), `auth.reset.public_base_url`, in-memory rate limits, and `password_reset_enabled` on `/server-info`
- Web UI reset page (`/reset-password`) and Forgot password flow on the login screen (mobile opens the email link in a browser)
- Mobile login/register: when the server field has no scheme or port, probe `GET /api/v1/status` over HTTPS then HTTP, update the field with the resolved URL (TLS/certificate errors still select HTTPS; if both fail, default to HTTPS as before)
- Android release allows cleartext HTTP so the client can reach local/LAN instances without TLS; iOS `Info.plist` sets `NSAllowsLocalNetworking` for the same local-HTTP case

### Changed

- Server URL field hint and validation copy no longer require typing `https://`

### Fixed

- Comment/like lists no longer crash when showing users without an avatar (`UserAvatar` only sets `onBackgroundImageError` when a network image is present)

## [0.6.0] - 2026-08-06

### Added

- Per-user `profile` preferences (`last_sport_type`, `last_equipment_by_sport`): file driver `users/<nickname>/profile.yaml`, bbolt bucket `user_profiles`; `GET /api/v1/profile`
- Creating a workout without `equipment_ids` uses equipment from the user's profile for that `sport_type` (explicit `[]` still means none)
- Workout likes with counts and liker lists in the list/detail UI; local + federated `Like`/`Undo`, file `likes.yaml`, bbolt likes buckets, and `GET/POST/DELETE /api/v1/workouts/{id}/likes`
- Workout comments (add/list/delete) on own and others' workouts; file `comments.yaml`, bbolt comment buckets, `GET/POST/DELETE /api/v1/workouts/{id}/comments`, federated `Create`/`Delete` Note with `inReplyTo`, and comments snapshot on Workout Update; UI comment control next to likes

### Changed

- **Breaking:** `last_equipment_by_sport` removed from user identity responses (`/auth/me`, login, register); clients should use `GET /api/v1/profile`
- New workout form defaults sport and equipment from `GET /api/v1/profile` (no extra own-workouts list for last sport)
- `last_sport_type` is the sport of the chronologically newest workout (`start_date`, then id); refreshed asynchronously after create/update/delete, and once at the end of a Strava import job
- File-driver outbound Like activity ids are stored as YAML (`object_id` + `activity_id`); plain-text activity-id-only files are still read and migrate via inbox reconstruction

### Fixed

- Android/iOS: tighter workout like/comment bar spacing (no bottom pad; IconButton vertical pad 8 with shrinkWrap)
- Federated workout comment/like avatars: recompute `is_local` from handle domain (ignore origin flag), store/export public avatar URLs, lazy-cache remote author avatars when listing, and warm avatar cache from workout comment/like snapshots
- `grom migrate-storage` now copies workout likes (local, federated cache, and outbound Like activity ids) between `file` and `bbolt`
- `grom migrate-storage` copies workout comments (local, federated cache, and outbound Create Note activity ids) between `file` and `bbolt`

## [0.5.0] - 2026-08-04

### Added

- Edit workout: add and remove photos (original + preview); `POST/DELETE /api/v1/workouts/{id}/media` with federation Update delivery
- Creating a workout without `equipment_ids` (JSON omit or multipart field absent) copies equipment from the previous workout of the same `sport_type`; an explicit empty list keeps the workout without equipment. Strava ZIP import is unchanged
- Integration screen shows Strava export instructions with an inline link to the Strava download page before the import button
- Android: Health Sync + Google Drive workout import (Integration toggle, Home sync button, Drive folder picker); `POST /workouts` accepts `external_id`; `GET /workouts/external` checks duplicates

### Fixed

- Android workout list: add 4px gaps under the map preview and photo strip so media is not flush against the map or card edge

### Changed

- Workout photo picker is disabled once 20 photos are selected (create and edit)
- New workout form defaults the sport type to the user's most recent workout (falls back to Run when none exist)
- Create/update workout requests from the Flutter client always send `equipment_ids` (including `[]`) so clearing equipment is distinct from omitting the field
- Workout `device` from FIT tracks drops the word "Strava" (e.g. Strava-reexported "Strava Wahoo ELEMNT" → "Wahoo ELEMNT"); track files are left unchanged
- **Breaking:** workout metadata field `strava_activity_id` replaced by `external_id` (`name` + `id`); Strava bulk import sets `name` to `strava`. Recreate storage or re-import after upgrade (no migration).

## [0.4.1] - 2026-07-30

### Added

- Create-workout form auto-fills the name from the localized sport type and updates it when the sport changes until the user edits the name
- `POST /api/v1/workouts/parse-track` returns optional `name` extracted from GPX metadata/track name or FIT workout / sport profile name

## [0.4.0] - 2026-07-29

### Added

- About screen shows the author name, source code repository, and app license
- Workout detail speed chart (distance km × speed km/h) with tap tooltip, avg/max rows; `GET /api/v1/workouts/{id}/speed` (precomputed chart, up to 500 points) and `speed_max_kmh` on workout responses
- Workout detail heart-rate chart (distance km or elapsed minutes × bpm, red fill) with tap tooltip, avg/max rows; `GET /api/v1/workouts/{id}/heartrate` (precomputed chart, up to 500 points); `heart_rate_max` on workout list/detail responses; `heartrate-chart.json` / bbolt `heart_rate_charts` (+ federated buckets)
- Sport types Nordic Walk (`NordicWalk`) and Ice Hockey (`IceHockey`); sport categories Strength, Team, and Racket

### Changed

- Equipment type "Shoes" uses the Material Symbols `steps` icon instead of the walking-person glyph
- Sport type picker regrouped (strength / team / racket categories) with updated labels, order, colors, and icons; Strava import IDs unchanged for existing sports
- **Breaking:** bbolt speed/heart-rate chart bucket values use packed binary instead of JSON (file driver still stores `speed-chart.json` / `heartrate-chart.json`). Recreate the bbolt database or re-attach tracks after upgrade.
- **Breaking:** speed chart storage replaces full speed sidecars (`speed.yaml` / `speed.json`): pre-downsampled chart only (`speed-chart.json` blob on file driver; `speed_charts` / `fed_speed_charts` bbolt buckets on bbolt driver). Re-upload workouts with tracks after upgrade.
- Heart-rate chart without GPS omits `distance_m`; the UI uses minutes from the first HR sample
- Speed and heart-rate charts keep finite zero samples (`speed_kmh` / `heart_rate_bpm` >= 0); NaN/Inf are still dropped. Toggle via `tracks.SpeedChartZeroPolicy` / `tracks.HeartRateChartZeroPolicy` (`ChartZeroKeep` or `ChartZeroOmit`). Re-attach tracks to refresh stored charts.

### Fixed

- `grom migrate-storage` copies speed/heart-rate charts between file JSON blobs and bbolt binary buckets (alongside other metadata)
- `/api/docs` (no trailing slash) redirects to `/api/docs/` so Swagger UI assets load instead of a blank page
- bbolt: renaming a workout (e.g. edit that changes `start_date`) migrates speed/heart-rate chart buckets instead of deleting them; edit form preserves start-time seconds so silent truncations no longer trigger a rename

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

[Unreleased]: https://github.com/solargate/grom/compare/0.8.0...HEAD
[0.8.0]: https://github.com/solargate/grom/releases/tag/0.8.0
[0.7.3]: https://github.com/solargate/grom/releases/tag/0.7.3
[0.7.2]: https://github.com/solargate/grom/releases/tag/0.7.2
[0.7.1]: https://github.com/solargate/grom/releases/tag/0.7.1
[0.7.0]: https://github.com/solargate/grom/releases/tag/0.7.0
[0.6.0]: https://github.com/solargate/grom/releases/tag/0.6.0
[0.5.0]: https://github.com/solargate/grom/releases/tag/0.5.0
[0.4.1]: https://github.com/solargate/grom/releases/tag/0.4.1
[0.4.0]: https://github.com/solargate/grom/releases/tag/0.4.0
[0.3.0]: https://github.com/solargate/grom/releases/tag/0.3.0
[0.2.0]: https://github.com/solargate/grom/releases/tag/0.2.0
[0.1.0]: https://github.com/solargate/grom/releases/tag/0.1.0
