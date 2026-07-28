# Configuration

Grom is configured with a YAML file. By default it looks for `config.yaml` in the current working directory. Pass `--config` (or `-c`) to use another path.

- Example profiles: [`cmd/grom/config-examples/`](../../cmd/grom/config-examples/)
- Every field with comments: [`config.full.yaml`](../../cmd/grom/config-examples/config.full.yaml)

**Required:** `auth.jwt_secret` — a long random secret used to sign JWT access tokens.

## Common settings

| Area | What to set |
|------|-------------|
| `server.port` / `server.tls` | Listen ports and TLS mode (`off`, `static`, or `autocert`) |
| `storage.driver` / `location` / `temp_dir` | `file` (default) or `bbolt`; data root and temp dirs |
| `storage.bbolt.path` | Optional path to `grom.db` when using bbolt (default: `{location}/grom.db`) |
| `federation.enabled` / `federation.domain` | ActivityPub; requires HTTPS |
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
- Legacy configs with `server.tls.enabled: true` (and no `mode`) are treated as `mode: static`.

## Storage drivers

| Driver | Metadata | Charts (speed / heart rate) | Blobs (tracks, photos, avatars, keys) |
|--------|----------|-----------------------------|----------------------------------------|
| `file` (default) | YAML under `storage.location` | `speed-chart.json` and `heartrate-chart.json` in each workout dir | Same tree |
| `bbolt` | JSON in `{location}/grom.db` (or `storage.bbolt.path`) | JSON in bbolt buckets `speed_charts` / `fed_speed_charts` and `heart_rate_charts` / `fed_heart_rate_charts` (federated inbox) | Same filesystem layout under `storage.location` |

`postgres` is reserved in config but not implemented.

Migrate metadata between drivers (stop the server first; blobs are shared and not copied):

```bash
grom migrate-storage --config config.yaml --from file --to bbolt --verify
# then set storage.driver: bbolt and restart
grom migrate-storage --config config.yaml --from bbolt --to file --verify
```

Use `--dry-run` to count records without writing, and `--force` to overwrite an existing bbolt database.

## See also

- [Install and run](install.md)
- [Strava bulk import](../strava-bulk-import.md) (`storage.temp_dir`)
