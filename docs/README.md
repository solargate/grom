# Documentation

Grom is a self-hosted workout tracker with an optional ActivityPub federation layer. Record or import workouts, manage equipment, follow other athletes (local or federated), like and comment on their activities, and browse a social feed - all on infrastructure you control.

## I want to…

### User

| Goal | Page |
|------|------|
| See what the client can do (workouts, likes, comments, recording, equipment) | [User overview](user/overview.md) |
| Reset a forgotten password (when the operator enables email) | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Sign in / register when the instance enables captcha | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Use Grom in a browser (same UI as Android) | Open the server base URL after [install](admin/install.md); see [User overview](user/overview.md) |
| Import a Strava export (UI + how import works) | [Strava bulk import](integrations/strava-bulk-import.md) |
| Import Health Sync activities from Google Drive (Android) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Create API tokens for scripts and external apps | [Grom API tokens](user/grom-api-tokens.md) |

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
