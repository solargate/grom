# Install and run

Install Grom from a GitHub release for the quickest setup. The server binary includes the web UI. For public deployments, see [Configuration](configuration.md).

## Download from GitHub Releases

Pre-built packages are published with each [GitHub release](https://github.com/solargate/grom/releases):

| Asset | Contents |
|-------|----------|
| `grom-<version>-linux-amd64.tar.gz` | Server binary + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Server binary + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Server binary + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Server binary + `config-examples/` |

The same release also provides the Android client as `grom-<version>.apk`.

## Run

1. Download and extract the server archive for your operating system and CPU. Open a terminal in the extracted directory, where the binary and `config-examples/` are located.
2. Copy the local HTTP example to `config.yaml` and replace `auth.jwt_secret` with a long random secret.
3. Start the server with that file.

**Linux and macOS**

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# Edit config.yaml and set auth.jwt_secret before starting.
./grom --config config.yaml
```

**Windows PowerShell**

```powershell
Copy-Item config-examples/config.dev.notls.yaml config.yaml
# Edit config.yaml and set auth.jwt_secret before starting.
.\grom.exe --config config.yaml
```

Open `http://localhost:8080/` in a browser. Register an account and sign in. The example uses HTTP and is intended for a local trial. Configure [TLS](configuration.md#tls-profiles) before exposing the server publicly.

Without `--config` (or `-c`), Grom looks for `config.yaml` in the current working directory. Relative storage paths are resolved from the binary directory.

An Android client can connect to the same instance from a device that can reach the server. Enter the server’s address as seen from that device. See [User overview](../user/overview.md) for server URL selection and sign-in. Swagger UI is available at `/api/docs/` on the running server.

## Next steps

- Choose a TLS profile and storage driver (`bbolt` for normal installs; `file` is mainly for tests / tiny instances) — [Configuration](configuration.md)
- Back up your server data and test recovery — [Backup and restore](backup-restore.md)
- Optionally enable password reset email (`mailer` + `auth.reset.public_base_url`) — [Configuration — Mailer and password reset](configuration.md#mailer-and-password-reset)
- Optionally enable ALTCHA captcha on register/login/forgot (`auth.captcha.enabled`) — [Configuration — Auth captcha](configuration.md#auth-captcha-altcha)
- Generate self-signed certs for local HTTPS — `grom gencerts` (see TLS section in configuration)
- Product tour for the client — [User overview](../user/overview.md)
- API reference in the browser — `/api/docs/` on the running server

## Build from source

### Requirements

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (for the embedded web UI and Android builds)
- Python 3 (MkDocs docs site and `server-catalog.yaml` tooling; `make web` / `make android-*` regenerate the catalog)
- Make

### Build

From the repository root:

```bash
make grom    # swagger + Flutter web + Go binary → cmd/grom/grom
make test    # Go and Flutter tests
make cli     # Go binary only (no Flutter web rebuild)
make web     # Flutter web → internal/web/dist
```

Other useful targets: `make apidoc` (regenerate OpenAPI), `make catalog` (server catalog Dart), `make android-apk`, `make clean`.

After the build, enter `cmd/grom` and follow the [run steps](#run) using the binary and configuration examples in that directory.
