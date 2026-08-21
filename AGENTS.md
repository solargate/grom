# AGENTS.md — Project Grom

Guidance for AI coding agents working in this repository.

## What this project is

**Grom** is a self-hosted workout tracker with an optional ActivityPub federation layer. Users record or import workouts (GPX/FIT), manage equipment, follow other users (local or federated), and browse a social feed.

- **License:** GPL-3.0
- **Module:** `github.com/solargate/grom`
- **Version:** single source of truth in `VERSION` (injected into Go via ldflags on `internal/version.Version` and Flutter via `--build-name`; Android `versionCode` is `--build-number` / `BUILD_NUMBER`)
- **Changelog:** Keep a Changelog format in `CHANGELOG.md`; release CI copies the matching version section into the GitHub draft release body (omitting `### Google Play`) and uploads English Play Store "What's new" from that block

## High-level architecture

```
cmd/grom          → CLI (`cmd/grom/cmd/`: root starts server; gencerts; migrate-storage)
api/v1            → HTTP handlers (Gin), wiring, Swagger annotations
api/docs          → Generated swagger artifacts (do not edit by hand)
internal/*        → Domain logic, storage, federation, tracks, auth, config
ui/grom           → Flutter client (web + Android; web build is embedded)
internal/web/dist → Embedded Flutter web build (copied by `make web`)
```

**Request flow:** `cmd/grom` → `config.GetConfig` → `api/v1.RunRouter` → `App` (services) → `storage.Backend` → file stores / blob store.

**UI flow:** Flutter talks to `/api/v1/*` with Bearer JWT (session) or a scoped PAT on workout/equipment routes. The same process serves the web UI via `internal/web` (path URLs; unmatched routes serve the Flutter SPA, including `/reset-password`).

## Directory map

| Path | Role |
|------|------|
| `CHANGELOG.md` | User-facing release notes (`[Unreleased]` + version sections; `### Google Play` is Play Store copy) |
| `scripts/` | Release helpers (`changelog_notes.sh`); `server_catalog.py` validates `server-catalog.yaml` and generates the Flutter catalog |
| `server-catalog.yaml` | Approved public Grom instances compiled into the Android client |
| `cmd/grom/` | Main binary; Cobra commands in `cmd/`; example configs in `config-examples/` |
| `api/v1/` | Gin handlers, route registration, DTO/response types |
| `api/docs/` | `swag`-generated OpenAPI (`make apidoc`) |
| `internal/auth/` | JWT + password hashing; `AuthRequired` (JWT only) and `AuthAPI` (JWT or scoped PAT); reset under `reset/`; captcha under `captcha/`; PAT under `pat/` |
| `internal/mailer/` | Outbound email (`off` / `log` / `smtp` via go-mail) |
| `internal/config/` | Viper YAML config; global `config.Cfg` |
| `internal/logging/` | `slog` setup from `logging.level` / `logging.format` |
| `internal/users/` | User repository + models; per-user `Profile` (last sport / last equipment by sport / used sport types) |
| `internal/workouts/` | Workout service, validation, feed, media, track attach |
| `internal/equipment/` | Equipment CRUD; mileage recalc in `distance/` |
| `internal/social/` | Follow graph + delivery hooks |
| `internal/federation/` | ActivityPub inbox/outbox, delivery, keys, avatar cache |
| `internal/tracks/` | GPX/FIT parse, stats, export, simplify |
| `internal/storage/` | `Backend` interface; `file` and `bbolt` drivers (postgres not implemented) |
| `internal/storage/file/` | File driver (YAML metadata + FS blobs) |
| `internal/storage/bbolt/` | BBolt driver (JSON metadata in Bolt, blobs on FS) |
| `internal/storage/migrate/` | Metadata copy between `file` ↔ `bbolt` (includes charts, likes, comments, PAT, profiles) |
| `internal/storage/blob/` | Blob keys + FS blob store |
| `internal/integrations/strava/` | Strava ZIP bulk import jobs |
| `internal/avatars/` | Avatar processing |
| `internal/maprender/` | Static map preview rendering |
| `internal/server/` | HTTP/HTTPS listen (static TLS, autocert) |
| `internal/version/` | Product version string (ldflags from `VERSION`) |
| `internal/web/` | `embed` of Flutter web assets |
| `ui/grom/` | Flutter app (`lib/`, `test/`) |
| `testdata/` | Shared fixtures: `tracks/` (GPX/FIT), `integrations/health-sync/` |
| `docs/` | Human documentation (EN + RU/DE via `*.ru.md` / `*.de.md`); index in `docs/README.md` (also MkDocs homepage) — not the same as `api/docs/` |
| `mkdocs.yml` / `requirements.txt` | GitHub Pages docs site (Material + `mkdocs-static-i18n`) and catalog generator (PyYAML); `make docs` / `make catalog` use `.venv`. Pages workflow: `.github/workflows/pages.yml` |
| `PRIVACY.md` | Stub linking to `docs/privacy.md` and Pages `/privacy/` |

## Human documentation

- **Languages:** English is the canonical source under `docs/` (unsuffixed files). Russian and German are full translations via MkDocs Material + `mkdocs-static-i18n` (`docs_structure: suffix`: `page.ru.md`, `page.de.md`). EN is served at `/`; RU at `/ru/`; DE at `/de/`. Root `README.md` / `AGENTS.md` stay English-only.
- **Layout:** Root `README.md` is a short front door (what/why, hero screenshots, quick start, links). Details live under `docs/`:
  - `docs/about.md` — project mission and pillars (EN + RU/DE)
  - `docs/user/` — end-user / client tour (web + Android)
  - `docs/admin/` — install, configuration (TLS, storage, federation, logging)
  - `docs/integrations/` — third-party import guides (Strava ZIP, Health Sync + Google Drive)
  - `docs/screenshots/` — images for README and user docs (shared across locales)
  - `docs/privacy.md` — privacy policy (English only; official text; root `PRIVACY.md` is a stub). Do not add `privacy.ru.md` / `privacy.de.md` — all locales link the same EN page via fallback.
- **Index:** `docs/README.md` (short pitch, Features, screenshot teaser, then “I want to…”). MkDocs treats it as the site homepage (`index.html`). Do not add a parallel `docs/index.md`. Do not duplicate that TOC here; keep the feature list in sync with root `README.md` when capabilities change. Localized homepages: `README.ru.md`, `README.de.md`.
- **Pages:** published to `https://solargate.github.io/grom/` via MkDocs Material (`.github/workflows/pages.yml` on `docs/**` changes and on release). Local preview: `make docs` / `make docs-serve` (creates `.venv` from `requirements.txt`). Store listings should use `…/privacy/`.
- **When to update:** client/UI behavior → `docs/user/`; install, TLS, storage, federation, logging → `docs/admin/`; third-party imports → `docs/integrations/` (keep README to a brief quick start + links; do not re-expand long config tables into README). Touch `docs/README.md` and `mkdocs.yml` nav if you add/rename pages. Prefer GitHub blob/tree URLs (not `../` repo paths) for links that leave `docs/`, so they work on Pages.
- **i18n sync:** when you change an English `docs/**/*.md` page (except `privacy.md`), update the matching `*.ru.md` and `*.de.md` in the same change. New pages need EN + RU + DE (and `nav_translations` in `mkdocs.yml` if nav labels change). Do not translate config keys, API paths, or code identifiers. For Russian headings that are linked from other pages, set explicit English-style anchors (`## Заголовок {#english-slug}`) so cross-links stay stable (Cyrillic alone yields `_1`-style slugs).
- **Do not confuse** `docs/` (human markdown) with `api/docs/` (generated OpenAPI; regenerate with `make apidoc`). Runtime Swagger UI is `/api/docs`.

## Tech stack

**Backend**

- Go 1.26+, Gin, Viper, JWT (`golang-jwt/jwt/v5`), swag/gin-swagger, structured logging via `log/slog` (+ `samber/slog-gin` for HTTP)
- Track formats: `tkrajina/gpxgo`, `muktihari/fit`
- Storage today: filesystem (`storage.driver: file`) or hybrid bbolt (`storage.driver: bbolt` — JSON metadata in Bolt, blobs on FS). Config also names `postgres` but it is **not implemented** — do not pretend it works.

**Frontend**

- Flutter (SDK `>=3.4.0 <4.0.0`), Material, `flutter_map`, geolocator, foreground task (Android recording)
- Shipped clients: web (embedded) + Android. `ui/grom/ios/` is Flutter scaffolding, not a product target.
- Locales: EN / RU / DE via ARB + generated `l10n`

## Build, run, test

Prefer Makefile targets:

```bash
make grom          # swagger + flutter web + go build → cmd/grom/grom
make cli           # go build only
make apidoc        # regenerate api/docs from swag annotations in api/v1
make docs          # MkDocs Material site → site/; creates `.venv` from requirements.txt
make docs-serve    # mkdocs serve (live preview)
make catalog       # validate server-catalog.yaml and generate ui/grom/lib/generated/server_catalog.g.dart
make web           # catalog + flutter build web → internal/web/dist
make test          # go test ./... && flutter test
make test-go
make test-ui
make android-apk   # release APK
make android-aab   # release App Bundle (Play)
make gencerts IP=... DOMAIN=...
make clean
```

CI (`.github/workflows/verify.yml` → `checks.yml`) runs `go vet` / `go test`, `flutter analyze` / `flutter test`, `make apidoc` with `git diff --exit-code api/docs`, and `scripts/server_catalog.py generate` with `git diff --exit-code` on the generated Flutter catalog. Flutter version is pinned in the workflow files.

Run server (from `cmd/grom` or with absolute config path):

```bash
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
# or after build:
./grom --config config.yaml
```

TLS / federation / storage are documented in `docs/admin/configuration.md` (install in `docs/admin/install.md`). Federation **requires** `server.tls.mode` of `static` or `autocert` (not `off`).

Storage driver switch: stop the server, run `grom migrate-storage --from file --to bbolt` (or the reverse; `--dry-run` / `--verify` / `--force` as needed), then set `storage.driver` and restart.

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
- File driver: `internal/storage/file/`; bbolt driver: `internal/storage/bbolt/`.
- Auth middleware: `AuthRequired()` accepts session JWT only (rejects `grom_pat_…`). `AuthAPI(pat, scope)` accepts JWT **or** a PAT with that scope. Use `AuthAPI` only for workouts/equipment. Social, likes, comments, profile, integrations, and account routes stay `AuthRequired`.
- API changes: add/update swag comments on handlers, then run `make apidoc`.
- Shared `ErrorResponse`: keep a neutral schema `example` (e.g. `"bad request"`); put endpoint-specific meaning on `@Failure` descriptions, not per-status DTOs.
- JSON/form DTOs live next to handlers in `api/v1`; domain models live in `internal/<pkg>/model.go`.
- Logging: use `log/slog` via `internal/logging` (configured by `logging.level` / `logging.format`). Prefer structured attrs (`"workout_id", id`, `"err", err`), not interpolated/`fmt.Sprintf` message strings. Levels: DEBUG diagnostics; INFO lifecycle; WARN recoverable; ERROR failed ops. HTTP access logs go through `samber/slog-gin` in `api/v1/main.go` — do not restore `gin.Default()`. Do not add zap/zerolog/logrus; keep `log.Fatal` only for pre-slog bootstrap in `config.GetConfig`. Do **not** use `fmt.Print*` / `log.Print*` for server diagnostics. `fmt.Errorf` / `%w` for **returned errors** is fine and is not logging — reserve slog for side-effect diagnostics (especially when the error is swallowed or best-effort). CLI user-facing output in `cmd/grom` (`fmt.Printf` for migrate/gencerts/version) is separate from server logging.
- Tests: colocated `*_test.go`; use `testdata/` (`tracks/`, `integrations/health-sync/`) for fixtures. Run `go test ./...` after backend changes.

### Flutter (`ui/grom`)

- Structure: `pages/`, `widgets/`, `services/`, `models/`, `navigation/`, `platform/`
- Platform splits use stub/io/web files — recording, Google Drive, file download, Strava archive picker, share intent, server scheme probe. Preserve that pattern. Do not expand iOS beyond existing scaffolding.
- API access: `lib/api_request.dart` + auth/server storage helpers.
- User-facing strings: update ARB files under `lib/l10n/` (en/ru/de); do not hardcode UI copy when localization exists. Generated `app_localizations*.dart` are committed (`flutter: generate: true`). When adding sport/equipment types, also update the hand-written `sport_type_localizations.dart` / `equipment_type_localizations.dart` switches.
- After Flutter UI changes that ship in the server binary, regenerate web embed: `make web` (or full `make grom`).
- Tests: `ui/grom/test/`; run `make test-ui`. CI also runs `flutter analyze` in `ui/grom`.

### Commits / scope

- Match existing style: short, imperative commit subjects focused on why.
- Do not commit generated noise, local data, TLS material, or `.cursor/`.
- `internal/web/dist/*` is produced by the build; prefer regenerating via Makefile rather than hand-editing.
- User-visible changes: add a bullet under `CHANGELOG.md` → `[Unreleased]` (Added / Changed / Fixed / Security; call out **Breaking** for config, API, or storage). Prefix the bullet with **UI:** (Flutter, all clients), **Android:** (Android-only), **Server:**, **Docs:**, or **CI:**. If the release includes **UI:** or **Android:**, keep a short English `### Google Play` block (plain text, ≤500 characters) for the Play Store listing; omit that block when the release is Server/Docs/CI only (CI then uses a maintenance stub). Skip pure refactors, tests, and CI noise.

## Domain notes agents should respect

1. **Workouts** are the core entity: metadata + optional track blob + media + map preview. IDs are short (`workouts.WorkoutIDLength`); newly allocated IDs are unique across all local users on the instance. `GET /workouts` is a cursor page `{items, next_cursor, has_more}` (`scope=feed|own`, default `feed`; `limit` default 20, max 100) — not a bare array. Import dedup uses `external_id` (`name` + `id`); `GET /workouts/external` checks whether that pair already exists. Strava sets `name=strava`; Health Sync sets `name=health-sync/{source}`.
2. **Tracks:** parse/enrich via `internal/tracks`; attach through workout service, not by writing files from handlers alone. `POST /workouts/parse-track` is the client pre-create parse.
3. **Social feed** merges local workouts with federated inbox content (`workouts.FeedService` + federation adapters).
4. **Workout likes:** cannot like own workouts; API `GET/POST/DELETE /workouts/{id}/likes` (optional `owner` query like get workout). Responses expose `likes_count`, `liked_by_me`, `can_like`. Local likes via `workouts.LikesRepository`; file: `likes.yaml` per workout, federated cache/outbox under `federation/`; bbolt: `workout_likes` / `fed_workout_likes` / `like_activities`. `grom migrate-storage` copies local likes, federated like cache, and outbound Like activity ids between drivers. Federated like/unlike delivers ActivityPub `Like` / `Undo`; inbox applies remote likes and caches `likesCount` / `likedUsers` from Create objects. UI: `WorkoutLikeBar` on list cards and detail (likes left, comments right). JWT only (not PAT).
5. **Workout comments:** can comment on own and others' workouts; API `GET/POST /workouts/{id}/comments`, `DELETE /workouts/{id}/comments/{commentId}` (optional `owner`). Text max 1000 chars; empty rejected. Delete allowed for comment author or workout owner. Responses expose `comments_count`; list items include `can_delete`. Local via `workouts.CommentsRepository`; file: `comments.yaml` (`comments_num` + `comments[]` with `id`, `user`, `datetime`, `text`, `note_id`); federated cache/outbox under `federation/`; bbolt: `workout_comments` / `fed_workout_comments` / `comment_activities`. Federated comment delivers ActivityPub `Create`/`Note` with `inReplyTo`; delete delivers `Delete` Note (owner delete of remote comment notifies author). Workout Create/Update embeds `commentsCount` / `comments`. UI dialog for list/add/delete. JWT only (not PAT).
6. **Federation** (ActivityPub): WebFinger, actor, inbox/outbox, shared inbox under root paths (not only `/api/v1`). Delivery is async with retry workers. Keep HTTP signatures / actor URLs consistent with `federation.domain`.
7. **Strava import:** background jobs under `internal/integrations/strava`; column mapping and behavior are documented in `docs/integrations/strava-bulk-import.md`. JWT only (not PAT).
8. **Health Sync:** Android-only client import from Google Drive (`ui/grom/lib/services/health_sync_*`, `platform/google_drive_*`). There is **no** backend Health Sync package — workouts are created through the normal API with `external_id`. Sign-out or a server URL change turns Drive import off. Docs: `docs/integrations/health-sync-google-drive.md`.
9. **Avatars:** local users + federated author avatar cache; public federation avatar routes differ from authenticated API avatar routes.
10. **Speed chart:** pre-downsampled series (≤500 pts) written at track attach; `GET /workouts/{id}/speed` reads chart only. File driver: `speed-chart.json` blob (JSON for debuggability); bbolt driver: packed binary values in `speed_charts` / `fed_speed_charts` buckets (tracks/media stay on FS).
11. **Heart rate chart:** same pattern as speed (`heartrate-chart.json` on file; packed binary in bbolt `heart_rate_charts` / `fed_heart_rate_charts`); `GET /workouts/{id}/heartrate`; `distance_m` omitted without GPS; X axis is distance km or elapsed minutes from first HR sample.
12. **Password reset:** optional; enabled when `mailer.driver` is `log`/`smtp` and `auth.reset.public_base_url` is set. API `POST /auth/password/forgot` and `/auth/password/reset`; tokens in `reset_tokens.yaml` / bbolt `reset_tokens` (not migrated by `grom migrate-storage`). UI: Forgot password on login + web `/reset-password` (mobile opens email link in browser). `password_reset_enabled` on `/server-info`.
13. **Auth captcha (ALTCHA):** optional PoW via `auth.captcha.enabled` (default off). Protects register, login, and password forgot (not token reset). `GET /api/v1/captcha/challenge`; client sends `altcha` payload; `captcha_enabled` on `/server-info`. Implementation in `internal/auth/captcha/`.
14. **Personal access tokens (PAT):** scoped long-lived API tokens (`grom_pat_…`) for **own local** workouts/equipment only (not feed, not federated). `GET/POST /api/v1/auth/pat`, `DELETE /api/v1/auth/pat/{id}` (JWT only); `AuthAPI` on workout/equipment routes. Storage in `personal_access_tokens.yaml` / bbolt `personal_access_tokens` (copied by `grom migrate-storage`). Scopes: `workouts:read`, `workouts:write`, `equipment:read`, `equipment:write`. Default TTL 90 days, max 10 per user. UI: Integration → Grom tab. Docs: `docs/user/grom-api-tokens.md`.
15. **Account deletion:** `DELETE /api/v1/auth/me` (JWT + password body only; PAT rejected). Delivers ActivityPub `Delete` of the actor best-effort, then `storage.Backend.PurgeUser` (file: wipe `users/{nick}` + `users.yaml` + cross-user follows/likes/comments/inbox; bbolt: equivalent bucket purge + FS user dir). Inbox handles remote actor `Delete`. Nickname/email may be re-registered immediately. UI: profile menu → password dialog → goodbye → logout.
16. **User profile preferences:** `GET /profile` (JWT); `last_sport_type`, `last_equipment_by_sport`, and `used_sport_types` (unique sports from create/update, most recently used first; pruned when no remaining workouts use that type). File: `profile.yaml`; bbolt: `user_profiles`. Copied by `grom migrate-storage`.
17. **Equipment mileage:** cached `distance` is recalculated by `internal/equipment/distance` from the owner's workouts (including after Strava import). Do not set mileage only from the handler.
18. **Sport and equipment type catalogs** must stay in sync: Go `internal/workouts/sport_types.go` ↔ Flutter `lib/models/sport_types.dart` (plus ARB keys and `sport_type_localizations.dart`); equipment types `internal/equipment` ↔ `lib/models/equipment_types.dart` (plus `equipment_type_localizations.dart`).
19. **Approved servers:** `server-catalog.yaml` at the repo root is the source of truth; `scripts/server_catalog.py` validates it (https only, no ports, path allowed) and writes `ui/grom/lib/generated/server_catalog.g.dart`. Commit the generated file; `make catalog` / `make web` / `make android-*` regenerate it. CI diffs the Dart file. Do not parse the YAML in Flutter at runtime. Manual server URLs stay allowed; successful login/register remembers custom URLs locally. Docs: `docs/user/approved-servers.md`.

## Agent do / don't

**Do**

- Read nearby packages before changing interfaces (`Repository`, `Backend`, `blob.Store`).
- Preserve error-sentinel patterns and HTTP mapping in `api/v1`.
- Pick `AuthRequired` vs `AuthAPI` explicitly for new `/api/v1` routes (see coding conventions).
- Update config examples and `docs/admin/` when TLS/federation/storage/install behavior changes; keep root `README.md` short (link out, update quick start only if needed). Sync `*.ru.md` / `*.de.md` with English docs edits.
- Update `docs/user/` when client-facing flows or screens change in a way operators/users need to know (EN + RU + DE).
- Update `CHANGELOG.md` `[Unreleased]` for user-visible changes (component prefix; `### Google Play` when UI/Android).
- Add or extend tests for non-trivial logic (track stats, federation inbox, storage).
- Keep API and Flutter models aligned when changing JSON field names; keep sport/equipment catalogs in sync across Go and Flutter.

**Don't**

- Implement or claim support for `postgres` unless actually wiring a new driver behind `storage.Open`.
- Enable federation paths that assume HTTPS while leaving `tls.mode: off`.
- Hand-edit `api/docs/*` — regenerate with `make apidoc`.
- Hand-edit `ui/grom/lib/generated/server_catalog.g.dart` — regenerate with `make catalog`.
- Bypass auth middleware on protected `/api/v1` routes, or accept PAT on JWT-only routes (social, likes, comments, profile, integrations, account).
- Add a backend Health Sync package; keep that flow client-side on Android.
- Commit `cmd/grom/grom`, runtime `data/`, secrets, or personal config.
- Expand scope into unrelated refactors; match the request.

## Useful entry points

| Task | Start here |
|------|------------|
| Add API endpoint | `api/v1/app.go` (routes) + new/existing handler file + swag comments; choose `AuthRequired` vs `AuthAPI` |
| Change workout rules | `internal/workouts/` then handlers |
| Storage / on-disk layout | `internal/storage/`, `internal/storage/file/`, `internal/storage/bbolt/`, `internal/storage/keys/` |
| Storage migrate CLI | `cmd/grom/cmd/migrate_storage.go`, `internal/storage/migrate/` |
| Federation behavior | `internal/federation/`, `api/v1/federation_routes.go` |
| Track parsing/stats | `internal/tracks/` |
| Equipment mileage | `internal/equipment/distance/` |
| Flutter screen/API | `ui/grom/lib/pages/`, `api_request.dart` |
| Health Sync (Android) | `ui/grom/lib/services/health_sync_service.dart`, `platform/google_drive*.dart`; docs in `docs/integrations/health-sync-google-drive.md` |
| Sport / equipment types | `internal/workouts/sport_types.go`, `ui/grom/lib/models/sport_types.dart`; `internal/equipment`, `ui/grom/lib/models/equipment_types.dart` |
| Config / TLS listen | `internal/config/`, `internal/server/`; human docs in `docs/admin/configuration.md` |
| Password reset / mailer | `internal/auth/reset/`, `internal/mailer/`, `api/v1/auth_password.go`; docs in `docs/admin/configuration.md` |
| Auth captcha (ALTCHA) | `internal/auth/captcha/`, `api/v1/captcha.go`; Flutter `widgets/altcha_field.dart` |
| Personal access tokens | `internal/auth/pat/`, `api/v1/pat.go`, `internal/auth/middleware.go`; Flutter `pages/grom_api_tab.dart`; docs in `docs/user/grom-api-tokens.md` |
| Logging | `internal/logging/`, `logging:` in `cmd/grom/config-examples/` |
| Human docs | `docs/README.md` (index + Pages homepage), `docs/about.md`, `*.ru.md` / `*.de.md`, `docs/user/`, `docs/admin/`, `docs/integrations/`, `docs/privacy.md` (EN only); `mkdocs.yml` + `mkdocs-static-i18n`; keep root `README.md` short |
| Approved server catalog | `server-catalog.yaml`; `scripts/server_catalog.py`; Flutter `lib/server_catalog.dart` + `widgets/server_url_field.dart`; docs in `docs/user/approved-servers.md` |
| Changelog / Play notes | `CHANGELOG.md`; `scripts/changelog_notes.sh` (`github` strips `### Google Play`; `play` emits en-US "What's new") |
| Version bump / release | edit `VERSION`; move `CHANGELOG.md` `[Unreleased]` → `## [X.Y.Z] - YYYY-MM-DD` (keep `### Google Play` if UI/Android); update compare links; tag `X.Y.Z` on master (CI fills GitHub body from changelog minus Play block; Play "What's new" from `### Google Play` or a stub) |
