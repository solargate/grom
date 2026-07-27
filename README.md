# Project Grom

> ⚠️ **Warning — early development**
>
> Grom is under active development and is closer to an MVP than production software. It is not ready for production use.
>
> Configuration formats, APIs, storage layouts, and other interfaces may and will change without a stable migration path. Expect bugs. Use at your own risk.

Self-hosted workout tracker with an optional ActivityPub federation layer. Record or import workouts, manage equipment, follow other athletes (local or federated), and browse a social feed — all on infrastructure you control.

The server is a single Go binary; the Flutter client ships as an embedded web UI and as an Android app.

![Grom workout feed on Android](docs/screenshots/workout-list.jpg)

## Features

- **Workouts** — create and edit activities with stats, notes, media, and map previews
- **GPS tracks** — import GPX and FIT; live recording on Android
- **Equipment** — track bikes, shoes, and other gear linked to workouts
- **Social feed** — follow users and see their workouts in one timeline
- **Federation** — optional ActivityPub so instances can follow each other across the network
- **Strava import** — bulk-import a Strava data export ZIP
- **Clients** — web UI served by the same process; Android APK via the build
- **Locales** — English, Russian, and German in the Flutter UI

## Quick start

```bash
make grom    # swagger + Flutter web + Go binary → cmd/grom/grom
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
```

Set `auth.jwt_secret` in your config (required). Example profiles: `cmd/grom/config-examples/`.

## Documentation

- **[Docs index](docs/README.md)** — user and admin guides
- [User overview](docs/user/overview.md) — client screens (workouts, recording, equipment)
- [Install and run](docs/admin/install.md) — build and start the server
- [Configuration](docs/admin/configuration.md) — TLS, storage, federation, logging

## License

GPL-3.0

## Donations

If you’d like to support development, see [DONATIONS.md](DONATIONS.md).
