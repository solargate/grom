# Documentation

English documentation for Grom. The [root README](../README.md) is a short entry point; details live here.

## I want to…

### User

| Goal | Page |
|------|------|
| See what the client can do (workouts, likes, comments, recording, equipment) | [User overview](user/overview.md) |
| Reset a forgotten password (when the operator enables email) | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Sign in / register when the instance enables captcha | [User overview — Sign-in and password reset](user/overview.md#sign-in-and-password-reset) |
| Use Grom in a browser (same UI as Android) | Open the server base URL after [install](admin/install.md); see [User overview](user/overview.md) |
| Import a Strava export (UI + how import works) | [Strava bulk import](strava-bulk-import.md) |
| Import Health Sync activities from Google Drive (Android) | [Health Sync + Google Drive](health-sync-google-drive.md) |

### Admin

| Goal | Page |
|------|------|
| Build and run the server | [Install and run](admin/install.md) |
| Configure TLS, storage, federation, logging, mailer / password reset, captcha | [Configuration](admin/configuration.md) |

### Reference

| Topic | Page |
|-------|------|
| Strava ZIP column mapping and import behavior | [Strava bulk import](strava-bulk-import.md) |
| Health Sync Google Drive import (Android client) | [Health Sync + Google Drive](health-sync-google-drive.md) |
| Full annotated config | [`config.full.yaml`](../cmd/grom/config-examples/config.full.yaml) |
| HTTP API (Swagger UI) | On a running server: `/api/docs/` (e.g. `http://localhost:8080/api/docs/`); sources under [`api/docs/`](../api/docs/) |
