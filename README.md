# Project Grom

> ⚠️ **Warning — early development**
>
> Grom is under active development and is closer to an MVP than production software. It is not ready for production use.
>
> Configuration formats, APIs, storage layouts, and other interfaces may and will change without a stable migration path. Expect bugs. Use at your own risk.

> **Grom — a free workout ecosystem: your server, your communities, open code.**

Grom aims to be a fully free ecosystem for athletes and enthusiasts: run a server for yourself or any community, keep workouts on infrastructure you control, and optionally share activities across communities through federation. It stays open to integrations with available sports services and formats — and remains open source so anyone can help shape it. Longer version: [About the project](docs/about.md).

Self-hosted workout tracker with an optional ActivityPub federation layer. Record or import workouts, manage equipment, follow other athletes (local or federated), like and comment on their activities, and browse a social feed — all on infrastructure you control.

The server is a single Go binary; the Flutter client ships as an embedded web UI and as an Android app.

## Features

- **Workouts** — create and edit activities with stats, notes, media, and map previews
- **GPS tracks** — import GPX and FIT; live recording on Android
- **Equipment** — track bikes, shoes, and other gear linked to workouts
- **Social feed** — follow users and see their workouts in one timeline
- **Workout likes** — like others’ activities, see counts and who liked; federated `Like` / `Undo` when ActivityPub is enabled
- **Workout comments** — comment on own or others’ activities (add/list/delete); federated `Create`/`Delete` Note when ActivityPub is enabled
- **Federation** — optional ActivityPub so instances can follow each other across the network
- **Strava import** — bulk-import a Strava data export ZIP; on Android, optional BYO Strava API sync for recent activities
- **Import tracks** — multi-select GPX/FIT via the system file picker (web and Android)
- **API tokens** — scoped personal access tokens (`grom_pat_…`) for workouts and equipment
- **Clients** — same Flutter UI in the browser (served by the server) and as an Android APK
- **Locales** — English, Russian, and German in the Flutter UI

<p align="center">
  <img src="docs/screenshots/workout-list.jpg" width="250" alt="Workout list" />
  <img src="docs/screenshots/workout-record.jpg" width="250" alt="Live recording" />
  <img src="docs/screenshots/equipment.jpg" width="250" alt="Equipment" />
</p>

## Quick start

Download the server archive for your platform from [GitHub Releases](https://github.com/solargate/grom/releases) and unpack it. The binary includes the web UI.

From the extracted directory, copy `config-examples/config.dev.notls.yaml` to `config.yaml` and replace `auth.jwt_secret` with a long random secret. Run `./grom --config config.yaml` on Linux or macOS, or `.\grom.exe --config config.yaml` in Windows PowerShell. Open `http://localhost:8080/` to register and sign in.

See [Install and run](docs/admin/install.md) for platform-specific steps, public deployment settings, and building from source.

## Documentation

**On the website:** [Docs](https://grom.solargate.team/) · [About](https://grom.solargate.team/about/) · [User overview](https://grom.solargate.team/user/overview/) · [Delete account](https://grom.solargate.team/user/delete-account/) · [Install](https://grom.solargate.team/admin/install/) · [Configuration](https://grom.solargate.team/admin/configuration/) · [Privacy](https://grom.solargate.team/privacy/)

**In the repository** (for contributors):

- **[Docs index](docs/README.md)** — user and admin guides
- [About the project](docs/about.md) — why Grom exists and what guides it
- [User overview](docs/user/overview.md) — client screens (workouts, likes, comments, password reset, captcha, recording, equipment, track import)
- [Approved Grom servers](docs/user/approved-servers.md) — Android server picker and how to list a public instance
- [Delete your Grom account](docs/user/delete-account.md) — how to permanently delete an account (web or Android)
- [Grom API tokens](docs/user/grom-api-tokens.md) — personal access tokens for scripts and external apps
- [Import tracks](docs/integrations/import-tracks.md) — GPX/FIT via system file picker
- [Strava API import](docs/integrations/strava-api-import.md) — Android BYO Strava OAuth sync
- [Install and run](docs/admin/install.md) — install a release or build from source
- [Configuration](docs/admin/configuration.md) — TLS, storage, federation, logging, mailer / password reset, captcha
- [Privacy policy](docs/privacy.md)
- API docs — `/api/docs/` on a running server (OpenAPI sources in [`api/docs/`](api/docs/))

## License

GPL-3.0

## Donations

If you’d like to support development, see [DONATIONS.md](DONATIONS.md).
