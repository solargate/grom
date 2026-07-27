# AGENTS.md — Project Grom

Guidance for AI coding agents working in this repository.

## What this project is

**Grom** is a self-hosted workout tracker with an optional ActivityPub federation layer. Users record or import workouts (GPX/FIT), manage equipment, follow other users (local or federated), and browse a social feed.

- **License:** GPL-3.0
- **Module:** `github.com/solargate/grom`
- **Version:** single source of truth in `VERSION` (injected into Go via ldflags and Flutter via `--build-name`)
- **Changelog:** Keep a Changelog format in `CHANGELOG.md`; release CI copies the matching version section into the GitHub draft release body

## High-level architecture

```
cmd/grom          → CLI entrypoint (Cobra: root starts server; gencerts subcommand)
api/v1            → HTTP handlers (Gin), wiring, Swagger annotations
api/docs          → Generated swagger artifacts (do not edit by hand)
internal/*        → Domain logic, storage, federation, tracks, auth, config
ui/grom           → Flutter client (web + Android; web build is embedded)
internal/web/dist → Embedded Flutter web build (copied by `make web`)
```

**Request flow:** `cmd/grom` → `config.GetConfig` → `api/v1.RunRouter` → `App` (services) → `storage.Backend` → file stores / blob store.

**UI flow:** Flutter app talks to `/api/v1/*` with Bearer JWT. Web UI is served from the same process via `internal/web`.

## Directory map

| Path | Role |
|------|------|
| `CHANGELOG.md` | User-facing release notes (`[Unreleased]` + version sections) |
| `cmd/grom/` | Main binary; example configs in `config-examples/` |
| `api/v1/` | Gin handlers, route registration, DTO/response types |
| `api/docs/` | `swag`-generated OpenAPI (`make doc`) |
| `internal/auth/` | JWT + password hashing + `AuthRequired` middleware |
| `internal/config/` | Viper YAML config; global `config.Cfg` |
| `internal/logging/` | `slog` setup from `logging.level` / `logging.format` |
| `internal/users/` | User repository + models |
| `internal/workouts/` | Workout service, validation, feed, media, track attach |
| `internal/equipment/` | Equipment CRUD repository |
| `internal/social/` | Follow graph + delivery hooks |
| `internal/federation/` | ActivityPub inbox/outbox, delivery, keys, avatar cache |
| `internal/tracks/` | GPX/FIT parse, stats, export, simplify |
| `internal/storage/` | `Backend` interface; `file` and `bbolt` drivers (postgres not implemented) |
| `internal/storage/migrate/` | Metadata copy between `file` ↔ `bbolt` |
| `internal/storage/blob/` | Blob keys + FS blob store |
| `internal/integrations/strava/` | Strava ZIP bulk import jobs |
| `internal/avatars/` | Avatar processing |
| `internal/maprender/` | Static map preview rendering |
| `internal/server/` | HTTP/HTTPS listen (static TLS, autocert) |
| `internal/web/` | `embed` of Flutter web assets |
| `ui/grom/` | Flutter app (`lib/`, `test/`) |
| `testdata/` (`testdata/tracks/` for GPX/FIT) | Shared fixtures for Go and Flutter tests |
| `docs/` | Human documentation (English); index in `docs/README.md` — not the same as `api/docs/` |

## Human documentation

- **Language:** English.
- **Layout:** Root `README.md` is a short front door (what/why, hero screenshots, quick start, links). Details live under `docs/`:
  - `docs/user/` — end-user / client tour (web + Android)
  - `docs/admin/` — install, configuration (TLS, storage, federation, logging)
  - `docs/strava-bulk-import.md` — Strava ZIP reference (column mapping, behavior)
  - `docs/screenshots/` — images for README and user docs
- **Index:** `docs/README.md` (“I want to…”). Do not duplicate that TOC here.
- **When to update:** client/UI behavior → `docs/user/`; install, TLS, storage, federation, logging → `docs/admin/` (keep README to a brief quick start + links; do not re-expand long config tables into README). Touch `docs/README.md` if you add/rename pages.
- **Do not confuse** `docs/` (human markdown) with `api/docs/` (generated OpenAPI; regenerate with `make doc`). Runtime Swagger UI is `/api/docs`.

## Tech stack

**Backend**

- Go 1.26+, Gin, Viper, JWT (`golang-jwt/jwt/v5`), swag/gin-swagger, structured logging via `log/slog` (+ `samber/slog-gin` for HTTP)
- Track formats: `tkrajina/gpxgo`, `muktihari/fit`
- Storage today: filesystem (`storage.driver: file`) or hybrid bbolt (`storage.driver: bbolt` — JSON metadata in Bolt, blobs on FS). Config also names `postgres` but it is **not implemented** — do not pretend it works.

**Frontend**

- Flutter (SDK `>=3.4.0 <4.0.0`), Material, `flutter_map`, geolocator, foreground task (Android recording)
- Locales: EN / RU / DE via ARB + generated `l10n`

## Build, run, test

Prefer Makefile targets:

```bash
make grom          # swagger + flutter web + go build → cmd/grom/grom
make cli           # go build only
make doc           # regenerate api/docs from swag annotations in api/v1
make web           # flutter build web → copy into internal/web/dist
make test          # go test ./... && flutter test
make test-go
make test-ui
make android-apk   # release APK
make gencerts IP=... DOMAIN=...
make clean
```

Run server (from `cmd/grom` or with absolute config path):

```bash
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
# or after build:
./grom --config config.yaml
```

TLS / federation / storage are documented in `docs/admin/configuration.md` (install in `docs/admin/install.md`). Federation **requires** `server.tls.mode` of `static` or `autocert` (not `off`).

## Configuration rules

- Required: `auth.jwt_secret`
- Defaults applied in `config.FinalizeConfig` (ports, JWT TTL, delivery workers, storage paths, logging level/format)
- Legacy `data:` YAML keys still map onto `storage:` location/temp_dir
- Config examples live under `cmd/grom/config-examples/` — keep them in sync when adding config fields
- Runtime data dirs (`cmd/grom/data`, `tmp`, local `config.yaml`, built binary) are gitignored; do not commit secrets or user data

## Coding conventions

### Go

- Package layout: domain packages under `internal/` expose **repository interfaces** and sentinel errors (`ErrWorkoutNotFound`, etc.). Handlers in `api/v1` map `errors.Is` to HTTP status codes.
- Prefer small, focused files; keep handlers thin — validation and persistence belong in domain packages.
- New storage features must go through `storage.Backend` / blob keys (`internal/storage/keys`), not ad-hoc paths in handlers.
- File driver implementation: `internal/storage/file/`.
- API changes: add/update swag comments on handlers, then run `make doc`.
- JSON/form DTOs live next to handlers in `api/v1`; domain models live in `internal/<pkg>/model.go`.
- Logging: use `log/slog` via `internal/logging` (configured by `logging.level` / `logging.format`). Prefer structured attrs (`"workout_id", id`, `"err", err`), not `fmt`-style messages. Levels: DEBUG diagnostics; INFO lifecycle; WARN recoverable; ERROR failed ops. HTTP access logs go through `samber/slog-gin` in `api/v1/main.go` — do not restore `gin.Default()`. Do not add zap/zerolog/logrus; keep `log.Fatal` only for pre-slog bootstrap in `config.GetConfig`.
- Tests: colocated `*_test.go`; use `testdata/` (tracks under `testdata/tracks/`) for binary fixtures. Run `go test ./...` after backend changes.

### Flutter (`ui/grom`)

- Structure: `pages/`, `widgets/`, `services/`, `models/`, `navigation/`, `platform/`
- Platform splits use stub/io/web files (e.g. track recording store/foreground) — preserve that pattern when adding platform-specific behavior.
- API access: `lib/api_request.dart` + auth/server storage helpers.
- User-facing strings: update ARB files under `lib/l10n/` (en/ru/de); do not hardcode UI copy when localization exists.
- After Flutter UI changes that ship in the server binary, regenerate web embed: `make web` (or full `make grom`).
- Tests: `ui/grom/test/`; run `make test-ui`.

### Commits / scope

- Match existing style: short, imperative commit subjects focused on why.
- Do not commit generated noise, local data, TLS material, or `.cursor/`.
- `internal/web/dist/*` is produced by the build; prefer regenerating via Makefile rather than hand-editing.
- User-visible changes: add a bullet under `CHANGELOG.md` → `[Unreleased]` (Added / Changed / Fixed / Security; call out **Breaking** for config, API, or storage). Skip pure refactors, tests, and CI noise.

## Domain notes agents should respect

1. **Workouts** are the core entity: metadata + optional track blob + media + map preview. IDs are short (`workouts.WorkoutIDLength`); newly allocated IDs are unique across all local users on the instance.
2. **Tracks:** parse/enrich via `internal/tracks`; attach through workout service, not by writing files from handlers alone.
3. **Social feed** merges local workouts with federated inbox content (`workouts.FeedService` + federation adapters).
4. **Federation** (ActivityPub): WebFinger, actor, inbox/outbox, shared inbox under root paths (not only `/api/v1`). Delivery is async with retry workers. Keep HTTP signatures / actor URLs consistent with `federation.domain`.
5. **Strava import:** background jobs under `internal/integrations/strava`; column mapping and behavior are documented in `docs/strava-bulk-import.md`.
6. **Avatars:** local users + federated author avatar cache; public federation avatar routes differ from authenticated API avatar routes.

## Agent do / don't

**Do**

- Read nearby packages before changing interfaces (`Repository`, `Backend`, `blob.Store`).
- Preserve error-sentinel patterns and HTTP mapping in `api/v1`.
- Update config examples and `docs/admin/` when TLS/federation/storage/install behavior changes; keep root `README.md` short (link out, update quick start only if needed).
- Update `docs/user/` when client-facing flows or screens change in a way operators/users need to know.
- Update `CHANGELOG.md` `[Unreleased]` for user-visible changes.
- Add or extend tests for non-trivial logic (track stats, federation inbox, storage).
- Keep API and Flutter models aligned when changing JSON field names.

**Don't**

- Implement or claim support for `postgres` unless actually wiring a new driver behind `storage.Open`.
- Enable federation paths that assume HTTPS while leaving `tls.mode: off`.
- Hand-edit `api/docs/*` — regenerate with `make doc`.
- Bypass auth middleware on protected `/api/v1` routes.
- Commit `cmd/grom/grom`, runtime `data/`, secrets, or personal config.
- Expand scope into unrelated refactors; match the request.

## Useful entry points

| Task | Start here |
|------|------------|
| Add API endpoint | `api/v1/app.go` (routes) + new/existing handler file + swag comments |
| Change workout rules | `internal/workouts/` then handlers |
| Storage / on-disk layout | `internal/storage/`, `internal/storage/file/`, `internal/storage/keys/` |
| Federation behavior | `internal/federation/`, `api/v1/federation_routes.go` |
| Track parsing/stats | `internal/tracks/` |
| Flutter screen/API | `ui/grom/lib/pages/`, `api_request.dart` |
| Config / TLS listen | `internal/config/`, `internal/server/`; human docs in `docs/admin/configuration.md` |
| Logging | `internal/logging/`, `logging:` in `cmd/grom/config-examples/` |
| Human docs | `docs/README.md` (index), `docs/user/`, `docs/admin/`; keep `README.md` short |
| Version bump / release | edit `VERSION`; move `CHANGELOG.md` `[Unreleased]` → `## [X.Y.Z] - YYYY-MM-DD`; update compare links; tag `X.Y.Z` on master (CI fills release body from changelog) |
