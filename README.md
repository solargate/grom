# Project Grom

> ⚠️ **Warning — early development**
>
> Grom is under active development and is closer to an MVP than production software. It is not ready for production use.
>
> Configuration formats, APIs, storage layouts, and other interfaces may and will change without a stable migration path. Expect bugs. Use at your own risk.

Self-hosted workout tracker with an optional ActivityPub federation layer. Record or import workouts, manage equipment, follow other athletes (local or federated), and browse a social feed — all on infrastructure you control.

The server is a single Go binary; the Flutter client ships as an embedded web UI and as an Android app.

## Features

- **Workouts** — create and edit activities with stats, notes, media, and map previews
- **GPS tracks** — import GPX and FIT; live recording on Android
- **Equipment** — track bikes, shoes, and other gear linked to workouts
- **Social feed** — follow users and see their workouts in one timeline
- **Federation** — optional ActivityPub so instances can follow each other across the network
- **Strava import** — bulk-import a Strava data export ZIP (activities, tracks, photos, equipment)
- **Clients** — web UI served by the same process; Android APK available via the build
- **Locales** — English, Russian, and German in the Flutter UI

## Configuration

Grom is configured with a YAML file. By default it looks for `config.yaml` in the same directory as the server (the current working directory). Pass `-config` to use another path. Example profiles live in `cmd/grom/config-examples/`. For every field with comments, see `config.full.yaml`.

**Required:** `auth.jwt_secret` — a long random secret used to sign JWT access tokens.

Other common knobs:

| Area | What to set |
|------|-------------|
| `server.port` / `server.tls` | Listen ports and TLS mode (`off`, `static`, or `autocert`) |
| `storage.location` / `storage.temp_dir` | Data and temp directories (filesystem driver only today) |
| `federation.enabled` / `federation.domain` | ActivityPub; requires HTTPS |

### TLS profiles

| Profile | Config file | `tls.mode` | Federation |
|---------|-------------|------------|------------|
| Dev, HTTP only | `config.dev.notls.yaml` | `off` | disabled |
| Dev, self-signed TLS | `config.dev.tls.yaml` | `static` | enabled |
| Prod, HTTP only | `config.prod.notls.yaml` | `off` | disabled |
| Prod, Let's Encrypt | `config.prod.tls.yaml` | `autocert` | enabled |

**Dev with static TLS** — generate certificates, then run:

```bash
cd cmd/grom
go run . gencerts -ip 192.168.1.251 -domain 192.168.1.251
go run . -config config.dev.tls.yaml
```

For federation between local instances, set `federation.tls_insecure_skip_verify: true` and optionally `federation.ca_cert_file` to trust your dev CA.

**Production with autocert** — needs a public DNS name in `federation.domain` (hostname only), ports **80** and **443** reachable from the internet, and persistent `storage.location` (ACME cache under `{storage.location}/acme-cache`):

```bash
go run . -config config.prod.tls.yaml
```

Notes:

- Federation requires HTTPS (`tls.mode: static` or `autocert`). It cannot run with `tls.mode: off`.
- Legacy configs with `server.tls.enabled: true` (and no `mode`) are treated as `mode: static`.

## Build and run

```bash
make grom    # swagger + Flutter web + Go binary → cmd/grom/grom
make test    # Go and Flutter tests
```

```bash
cd cmd/grom && go run . -config config-examples/config.dev.notls.yaml
# or after build, with config.yaml next to the binary:
./grom
```

## License

GPL-3.0

## Donations

If you’d like to support development, see [DONATIONS.md](DONATIONS.md).
