# Configuration

Grom is configured with a YAML file. By default it looks for `config.yaml` in the current working directory. Pass `--config` (or `-c`) to use another path.

- Example profiles: [`cmd/grom/config-examples/`](../../cmd/grom/config-examples/)
- Every field with comments: [`config.full.yaml`](../../cmd/grom/config-examples/config.full.yaml)

**Required:** `auth.jwt_secret` — a long random secret used to sign JWT access tokens.

## Common settings

| Area | What to set |
|------|-------------|
| `server.port` / `server.tls` | Listen ports and TLS mode (`off`, `static`, or `autocert`) |
| `storage.driver` / `location` / `temp_dir` | `file` (default; tests / tiny instances) or `bbolt` (recommended for normal installs); data root and temp dirs |
| `storage.bbolt.path` | Optional path to `grom.db` when using bbolt (default: `{location}/grom.db`) |
| `federation.enabled` / `federation.domain` | ActivityPub; requires HTTPS |
| `auth.reset` / `mailer` | Password reset email (`public_base_url`, SMTP or log driver) |
| `logging.level` / `logging.format` | `debug`/`info`/`warn`/`error`; `text` (dev) or `json` (prod). Defaults: `info` + `json` |

Relative paths in `storage.*`, `server.tls.cert_file` / `key_file`, `server.tls.autocert.cache_dir`, and `federation.ca_cert_file` are resolved against the directory of the `grom` binary (absolute paths are used as-is).

## TLS profiles

| Profile | Config file | `tls.mode` | Federation |
|---------|-------------|------------|------------|
| Dev, HTTP only | `config.dev.notls.yaml` | `off` | disabled |
| Dev, self-signed TLS | `config.dev.tls.yaml` | `static` | enabled |
| Prod, HTTP only | `config.prod.notls.yaml` | `off` | disabled |
| Prod, Let's Encrypt | `config.prod.tls.yaml` | `autocert` | enabled |

**Dev with static TLS** — generate certificates, then run:

```bash
cd cmd/grom
go run . gencerts --ip 192.168.1.251 --domain 192.168.1.251
go run . --config config-examples/config.dev.tls.yaml
```

For federation between local instances, set `federation.tls_insecure_skip_verify: true` and optionally `federation.ca_cert_file` to trust your dev CA.

**Production with autocert** — needs a public DNS name in `federation.domain` (hostname only), ports **80** and **443** reachable from the internet. ACME certificates are cached under `acme-cache` next to the grom binary by default (override with `server.tls.autocert.cache_dir`; use an absolute path if the binary lives under a system directory like `/usr/bin`):

```bash
go run . --config config-examples/config.prod.tls.yaml
```

Notes:

- Federation requires HTTPS (`tls.mode: static` or `autocert`). It cannot run with `tls.mode: off`.
- With federation enabled, likes on remote workouts are delivered as ActivityPub `Like` / `Undo`, and comments as `Create`/`Delete` Note (`inReplyTo`); local workouts accept incoming likes and comments from other instances. Same-instance likes and comments work without federation.
- Legacy configs with `server.tls.enabled: true` (and no `mode`) are treated as `mode: static`.

## Storage drivers

`file` is fine for tests and very small instances (YAML metadata on disk is easy to inspect). For a normal or production install, prefer `bbolt` — metadata lives in a Bolt DB while tracks, photos, and other blobs stay on the filesystem.

| Driver | Metadata | Charts (speed / heart rate) | Workout likes / comments | Blobs (tracks, photos, avatars, keys) |
|--------|----------|-----------------------------|--------------------------|----------------------------------------|
| `file` (default) | YAML under `storage.location` (`users.yaml`, per-user `equipment.yaml`, `profile.yaml`, workout YAML, …) | `speed-chart.json` and `heartrate-chart.json` in each workout dir | `likes.yaml` and `comments.yaml` next to each local workout; federated like/comment cache and outbox activity IDs under the viewer’s `federation/` tree | Same tree |
| `bbolt` | JSON in `{location}/grom.db` (or `storage.bbolt.path`); includes `user_profiles` bucket for UI preferences | Packed binary values in bbolt buckets `speed_charts` / `fed_speed_charts` and `heart_rate_charts` / `fed_heart_rate_charts` (federated inbox) | Buckets `workout_likes`, `fed_workout_likes`, `like_activities`, `workout_comments`, `fed_workout_comments`, and `comment_activities` | Same filesystem layout under `storage.location` |

`postgres` is reserved in config but not implemented.

Migrate metadata between drivers (stop the server first; track/media/avatar blobs are shared and not copied). Speed and heart-rate charts are converted between file JSON blobs and bbolt binary buckets. Workout likes and comments (local, federated cache, and outbound activity ids) are copied so they remain readable after switching drivers:

```bash
grom migrate-storage --config config.yaml --from file --to bbolt --verify
# then set storage.driver: bbolt and restart
grom migrate-storage --config config.yaml --from bbolt --to file --verify
```

Use `--dry-run` to count records without writing, and `--force` to overwrite an existing bbolt database.

Password-reset tokens (`reset_tokens.yaml` / bbolt `reset_tokens`) are short-lived and are **not** copied by migrate-storage; in-flight reset links become invalid after a migrate.

## Mailer and password reset

Outbound email is optional. When `mailer.driver` is `off` (default), password reset is disabled and `GET /api/v1/server-info` reports `password_reset_enabled: false`.

| Setting | Purpose |
|---------|---------|
| `auth.reset.public_base_url` | Base URL embedded in reset links (no trailing slash). Required when mailer is on. |
| `auth.reset.token_ttl_minutes` | Token lifetime (default `60`) |
| `mailer.driver` | `off`, `log` (write to server log — useful in dev), or `smtp` |
| `mailer.from` | Sender address |
| `mailer.smtp.*` | External SMTP relay (`host`, `port`, `username`, `password`, `encryption`) |

There is **no** local MTA / `sendmail` dependency: the process speaks SMTP (via [go-mail](https://github.com/wneessen/go-mail)) to an external provider (Gmail app password, SES, Mailgun, etc.) or logs the message when `driver: log`.

Forgot-password requests are rate-limited in memory (per IP and per email). Behind a reverse proxy, configure Gin trusted proxies so `ClientIP` is correct.

Example (production SMTP on port 587):

```yaml
auth:
  jwt_secret: "..."
  reset:
    public_base_url: "https://grom.example.com"
mailer:
  driver: smtp
  from: "Grom <noreply@grom.example.com>"
  smtp:
    host: smtp.example.com
    port: 587
    username: "apikey"
    password: "secret"
    encryption: starttls
```

## See also

- [Install and run](install.md)
- [Strava bulk import](../strava-bulk-import.md) (`storage.temp_dir`)
