# Documentation

English documentation for Grom. The [root README](../README.md) is a short entry point; details live here.

## I want to…

### User

| Goal | Page |
|------|------|
| See what the client can do (workouts, recording, equipment) | [User overview](user/overview.md) |
| Import a Strava export (UI + how import works) | [Strava bulk import](strava-bulk-import.md) |

### Admin

| Goal | Page |
|------|------|
| Build and run the server | [Install and run](admin/install.md) |
| Configure TLS, storage, federation, logging | [Configuration](admin/configuration.md) |

### Reference

| Topic | Page |
|-------|------|
| Strava ZIP column mapping and import behavior | [Strava bulk import](strava-bulk-import.md) |
| Full annotated config | [`config.full.yaml`](../cmd/grom/config-examples/config.full.yaml) |
| HTTP API | Swagger UI when the server is running (`/swagger/index.html`), or generated artifacts under `api/docs/` |
