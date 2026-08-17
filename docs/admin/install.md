# Install and run

Install Grom from a GitHub release or build from source. For configuration details (TLS, storage, federation, mailer / password reset, captcha), see [Configuration](configuration.md).

## Download from GitHub Releases

Pre-built packages are published with each [GitHub release](https://github.com/solargate/grom/releases):

| Asset | Contents |
|-------|----------|
| `grom-<version>-linux-amd64.tar.gz` | Server binary + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Server binary + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Server binary + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Server binary + `config-examples/` |
| `grom-<version>.apk` | Android client |

Unpack the archive for your OS, copy a config from `config-examples/` (set `auth.jwt_secret`), then run the `grom` binary — see [Run](#run) below.

## Build from source

### Requirements

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (for the embedded web UI and Android builds)
- Make

### Build

From the repository root:

```bash
make grom    # swagger + Flutter web + Go binary → cmd/grom/grom
make test    # Go and Flutter tests
make cli     # Go binary only (no Flutter web rebuild)
make web     # Flutter web → internal/web/dist
```

Other useful targets: `make apidoc` (regenerate OpenAPI), `make android-apk`, `make clean`.

## Run

Example configs ship in release archives and in `cmd/grom/config-examples/` in the repo. The server looks for `config.yaml` in the current working directory unless you pass `--config` / `-c`.

**From a release package** (HTTP-only example):

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# edit config.yaml — set auth.jwt_secret
./grom --config config.yaml
```

**From source, HTTP only** (no TLS, federation off):

```bash
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
```

**After `make grom`**, with a config next to the binary:

```bash
cd cmd/grom
./grom --config config.yaml
# or rely on default config.yaml in the working directory:
./grom
```

Open the **web UI** in a browser at the server’s base URL (same Flutter client as Android — for example `http://localhost:8080/` with `config.dev.notls.yaml`). Register a user, then sign in. See [User overview](../user/overview.md).

The **Android** app (and later iOS) can connect to that same instance: enter the host on the login screen (scheme optional). Cleartext **HTTP is allowed for local/LAN** installs without TLS; use HTTPS when exposing the server beyond the local network. See [User overview](../user/overview.md) for how the client resolves `http` vs `https`.

**API docs (Swagger UI):** `http://<host>:<port>/api/docs/` (for example `http://localhost:8080/api/docs/`). Generated OpenAPI sources also live under [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) in the repository.

CLI help:

```bash
grom --help
grom --version
```

## Next steps

- Choose a TLS profile and storage driver (`bbolt` for normal installs; `file` is mainly for tests / tiny instances) — [Configuration](configuration.md)
- Optionally enable password reset email (`mailer` + `auth.reset.public_base_url`) — [Configuration — Mailer and password reset](configuration.md#mailer-and-password-reset)
- Optionally enable ALTCHA captcha on register/login/forgot (`auth.captcha.enabled`) — [Configuration — Auth captcha](configuration.md#auth-captcha-altcha)
- Generate self-signed certs for local HTTPS — `grom gencerts` (see TLS section in configuration)
- Product tour for the client — [User overview](../user/overview.md)
- API reference in the browser — `/api/docs/` on the running server
