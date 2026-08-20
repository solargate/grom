# Documentation

> **Grom — a free workout ecosystem: your server, your communities, open code.**

Grom aims to be a fully free ecosystem for athletes and enthusiasts: run a server for yourself or any community, keep workouts on infrastructure you control, and optionally share activities across communities through federation. It stays open to integrations with available sports services and formats — and remains open source so anyone can help shape it. Longer version: [About the project](about.md).

Grom is a self-hosted workout tracker with an optional ActivityPub federation layer. Record or import workouts, manage equipment, follow other athletes (local or federated), like and comment on their activities, and browse a social feed — all on infrastructure you control.

The server is a single Go binary; the Flutter client ships as an embedded web UI and as an Android app.

## Features

- **Workouts** — create and edit activities with stats, notes, media, and map previews
- **GPS tracks** — import GPX and FIT; live recording on Android
- **Equipment** — track bikes, shoes, and other gear linked to workouts
- **Social feed** — follow users and see their workouts in one timeline
- **Workout likes** — like others’ activities, see counts and who liked; federated `Like` / `Undo` when ActivityPub is enabled
- **Workout comments** — comment on own or others’ activities (add/list/delete); federated `Create`/`Delete` Note when ActivityPub is enabled
- **Federation** — optional ActivityPub so instances can follow each other across the network
- **Strava import** — bulk-import a Strava data export ZIP
- **Health Sync** — import activities from Google Drive on Android (Health Sync exports)
- **API tokens** — scoped personal access tokens (`grom_pat_…`) for workouts and equipment
- **Clients** — same Flutter UI in the browser (served by the server) and as an Android APK
- **Locales** — English, Russian, and German in the Flutter UI

<p align="center">
  <img src="screenshots/workout-list.jpg" width="250" alt="Workout list" />
  <img src="screenshots/workout-record.jpg" width="250" alt="Live recording" />
  <img src="screenshots/equipment.jpg" width="250" alt="Equipment" />
</p>

Screen-by-screen tour: [User overview](user/overview.md). Build and run: [Install and run](admin/install.md). Source and quick start: [repository README](https://github.com/solargate/grom#readme).

## I want to…

### About

| Goal | Page |
|------|------|
| Read the project mission and pillars | [About the project](about.md) |

### User

| Goal | Page |
|------|------|
| See what the client can do (workouts, likes, comments, recording, equipment) | [User overview](user/overview.md) |
| Reset a forgotten password (when the operator enables email) | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Sign in / register when the instance enables captcha | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Pick a public instance in the Android app, or list yours | [Approved Grom servers](user/approved-servers.md) |
| Use Grom in a browser (same UI as Android) | Open the server base URL after [install](admin/install.md); see [User overview](user/overview.md) |
| Import a Strava export (UI + how import works) | [Strava bulk import](integrations/strava-bulk-import.md) |
| Import Health Sync activities from Google Drive (Android) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Create API tokens for scripts and external apps | [Grom API tokens](user/grom-api-tokens.md) |
| Delete my account | [Delete your Grom account](user/delete-account.md) |

### Admin

| Goal | Page |
|------|------|
| Build and run the server | [Install and run](admin/install.md) |
| Configure TLS, storage, federation, logging, mailer / password reset, captcha | [Configuration](admin/configuration.md) |

### Reference

| Topic | Page |
|-------|------|
| Strava ZIP column mapping and import behavior | [Strava bulk import](integrations/strava-bulk-import.md) |
| Health Sync Google Drive import (Android client) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Full annotated config | [`config.full.yaml`](https://github.com/solargate/grom/blob/master/cmd/grom/config-examples/config.full.yaml) |
| HTTP API (Swagger UI) | On a running server: `/api/docs/` (e.g. `http://localhost:8080/api/docs/`); OpenAPI sources in [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) |
| Privacy policy | [Privacy Policy](privacy.md) |
