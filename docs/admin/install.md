# Install and run

Build and run the Grom server from source. For configuration details (TLS, storage, federation), see [Configuration](configuration.md).

## Requirements

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (for the embedded web UI and Android builds)
- Make

## Build

From the repository root:

```bash
make grom    # swagger + Flutter web + Go binary → cmd/grom/grom
make test    # Go and Flutter tests
make cli     # Go binary only (no Flutter web rebuild)
make web     # Flutter web → internal/web/dist
```

Other useful targets: `make doc` (regenerate OpenAPI), `make android-apk`, `make clean`.

## Run

Example configs live in `cmd/grom/config-examples/`. The server looks for `config.yaml` in the current working directory unless you pass `--config` / `-c`.

**Dev, HTTP only** (no TLS, federation off):

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

Open the web UI at the listen address from your config (see [Configuration](configuration.md)). Register a user in the UI, then sign in.

CLI help:

```bash
grom --help
grom --version
```

## Next steps

- Choose a TLS profile and storage driver — [Configuration](configuration.md)
- Generate self-signed certs for local HTTPS — `grom gencerts` (see TLS section in configuration)
- Product tour for the client — [User overview](../user/overview.md)
