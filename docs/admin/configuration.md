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
| `auth.captcha` | Optional ALTCHA PoW on register/login/forgot (`enabled`, optional `hmac_secret` / `cost` / `expires_seconds`) |
| `logging.level` / `logging.format` | `debug`/`info`/`warn`/`error`; `text` (dev) or `json` (prod). Defaults: `info` + `json`. Gin framework debug output (`[GIN-debug]`) is enabled only when `logging.level` is `debug`; otherwise Gin runs in release mode |

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

Migrate metadata between drivers (stop the server first; track/media/avatar blobs are shared and not copied). Speed and heart-rate charts are converted between file JSON blobs and bbolt binary buckets. Workout likes and comments (local, federated cache, and outbound activity ids) and personal access tokens are copied so they remain readable after switching drivers:

| Copied by `migrate-storage` | Not copied |
|-----------------------------|------------|
| Users, profiles, equipment, workouts | Password-reset tokens (short-lived; in-flight reset links become invalid) |
| Follows, federation followers, federated inbox authors/workouts | Blob files (tracks, photos, avatars, keys) — shared on disk |
| Local/federated likes and comments + outbound activity ids | Temp Strava jobs / in-memory captcha state |
| Speed/HR charts (format conversion) | |
| Personal access tokens (PAT) | |

```bash
grom migrate-storage --config config.yaml --from file --to bbolt --verify
# then set storage.driver: bbolt and restart
grom migrate-storage --config config.yaml --from bbolt --to file --verify
```

Use `--dry-run` to count records without writing, and `--force` to overwrite an existing bbolt database.

Password-reset tokens (`reset_tokens.yaml` / bbolt `reset_tokens`) are short-lived and are **not** copied by migrate-storage; in-flight reset links become invalid after a migrate. Legacy plain-text Like activity ids (without `object_id`) are reconstructed via federated inbox and, when `federation.domain` is set, local workout object URLs.

## Mailer and password reset

Outbound email is optional. When `mailer.driver` is `off` (default), password reset is disabled and `GET /api/v1/server-info` reports `password_reset_enabled: false`.

| Setting | Purpose |
|---------|---------|
| `auth.reset.public_base_url` | Base URL embedded in reset links (no trailing slash). Required when mailer is on. |
| `auth.reset.token_ttl_minutes` | Token lifetime (default `60`) |
| `mailer.driver` | `off`, `log` (write to server log — useful in dev), or `smtp` |
| `mailer.from` | Sender address |
| `mailer.smtp.host` / `port` | SMTP relay (required when `driver` is `smtp`). Common ports: `587` (STARTTLS), `465` (implicit TLS) |
| `mailer.smtp.username` / `password` | Optional SMTP credentials |
| `mailer.smtp.encryption` | `starttls` (default; also default for port 587), `tls` (default when port is `465`), or `none` |

There is **no** local MTA / `sendmail` dependency: the process speaks SMTP (via [go-mail](https://github.com/wneessen/go-mail)) to an external provider (Gmail app password, SES, Mailgun, etc.) or logs the message when `driver: log`.

Password-reset endpoints use an in-memory fixed-window rate limiter (15-minute window): forgot — 10 requests per client IP and 3 per email; confirm reset — 20 per client IP. Limits use Gin’s `ClientIP()` (honors `X-Forwarded-For` / `X-Real-IP` when present). Grom does not yet expose a trusted-proxies setting, so treat forwarded headers as untrusted unless your reverse proxy strips or overwrites them.

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

## Auth captcha (ALTCHA)

Optional self-hosted [ALTCHA](https://altcha.org/) proof-of-work captcha (no third-party service or API keys). When `auth.captcha.enabled` is `true`:

- **Protected:** `POST /api/v1/auth/register`, `/auth/login`, and `/auth/password/forgot` — request body must include a solved `altcha` payload (base64 JSON).
- **Not protected:** `POST /api/v1/auth/password/reset` (token from the email link is enough).
- **Challenge:** `GET /api/v1/captcha/challenge` returns a PoW challenge (`200`). When captcha is off → `404`; when the IP hits the challenge rate limit → `429` with `Retry-After`.
- **Client discovery:** `GET /api/v1/server-info` includes `captcha_enabled`. The Flutter web/Android UI shows an **I'm not a robot** checkbox and solves the challenge locally before submit.

| Setting | Purpose |
|---------|---------|
| `auth.captcha.enabled` | Require captcha on register/login/forgot. Default: `false`. |
| `auth.captcha.hmac_secret` | HMAC key for challenge signatures. Optional; when empty, `auth.jwt_secret` is used. Prefer a dedicated secret in production if you rotate JWT independently. |
| `auth.captcha.cost` | PBKDF2 iteration cost for PoW. Default: `1000`. Higher values cost more CPU on the client. |
| `auth.captcha.expires_seconds` | Challenge lifetime in seconds. Default: `300`. |

Challenge issuance is rate-limited in memory (60 requests per client IP per 15-minute window). Solved payloads are single-use until expiry (replay rejected). Rate limits and the replay store are process-local — they reset on restart and are not shared across multiple Grom processes; for multi-instance deploys, terminate TLS/rate-limit at a single reverse proxy or keep captcha off until sticky sessions / shared store exist.

Client IP resolution matches password reset: Gin’s `ClientIP()` (honors `X-Forwarded-For` / `X-Real-IP` when present). Without a trusted reverse proxy that overwrites those headers, treat them as untrusted.

Verification failures return `400` with messages such as `captcha is required`, `invalid captcha`, `captcha expired`, or `captcha already used`.

Example:

```yaml
auth:
  jwt_secret: "..."
  captcha:
    enabled: true
    # hmac_secret: "optional-separate-secret"
    cost: 1000
    expires_seconds: 300
```

## See also

- [Install and run](install.md)
- [Strava bulk import](../integrations/strava-bulk-import.md) (`storage.temp_dir`)
