# Project Grom

Self-hosted workout tracker with optional ActivityPub federation.

## TLS configuration

Grom supports four deployment profiles via `server.tls.mode`:

| Profile | Config file | `tls.mode` | Federation |
|---------|-------------|------------|------------|
| Dev, HTTP only | `config.dev.notls.yaml` | `off` | disabled |
| Dev, self-signed TLS | `config.dev.tls.yaml` | `static` | enabled |
| Prod, HTTP only | `config.prod.notls.yaml` | `off` | disabled |
| Prod, Let's Encrypt | `config.prod.tls.yaml` | `autocert` | enabled |

Example configs live in `cmd/grom/`.

### Dev with static TLS

Generate self-signed certificates before the first run:

```bash
cd cmd/grom
go run . gencerts -ip 192.168.1.251 -domain 192.168.1.251
go run . -config config.dev.tls.yaml
```

For federation between local instances, set `federation.tls_insecure_skip_verify: true`
and optionally `federation.ca_cert_file` to trust your dev CA.

### Production with autocert

Requirements:

- Public DNS name in `federation.domain` (hostname only, no port)
- Ports **80** and **443** reachable from the internet
- Persistent storage for `storage.location` (ACME cache is stored under `{storage.location}/acme-cache`)

```bash
go run . -config config.prod.tls.yaml
```

### Notes

- Federation requires HTTPS (`tls.mode: static` or `autocert`). It cannot be enabled with `tls.mode: off`.
- Legacy configs with `server.tls.enabled: true` (without `mode`) are treated as `mode: static`.
