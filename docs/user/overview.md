# User overview

Grom’s Flutter client runs as a **web UI** and as an **Android** app. The web UI is served by the same `grom` process: open the server’s base URL in a browser (for example `http://localhost:8080/` with the default dev config). The screens and flows match the Android app; live GPS recording is Android-only. UI strings are available in English, Russian, and German.

This page is a short tour of the main screens (screenshots below are from Android). Admin setup (install, config, TLS, federation) lives under [Admin docs](../README.md#admin). For the HTTP API, see Swagger at `/api/docs/` on a running server.

## Workouts

Your home feed lists activities with type, date, device, distance/time (and pace or elevation when relevant), plus a map preview when a GPS track is attached.

![Workout list on Android](../screenshots/workout-list.jpg)

You can create workouts manually, import GPX/FIT tracks, attach photos, and link equipment. On Android you can also record a live GPS track (see below).

## Live recording (Android)

Open **Add workout → Record** to track a route on the map with live duration, speed, and distance. Keep the recording notification while the session is active. Auto-pause can be toggled on or off.

![Recording a workout on Android](../screenshots/workout-record.jpg)

## Equipment

Manage bikes, shoes, and other gear. Items are grouped by category; distance totals accumulate as you use gear on workouts (including after Strava bulk import).

![Equipment list on Android](../screenshots/equipment.jpg)

## Social and federation

Follow other users on the same instance and browse a shared feed. When the operator enables ActivityPub federation, you can also follow athletes on other Grom instances (HTTPS required on the server).

## Strava import

From **Integration → Strava**, upload a [Strava bulk data export](https://support.strava.com/hc/en-us/articles/216918437-Exporting-your-Data-and-Bulk-Export) ZIP. Grom imports activities, tracks, photos, and equipment. Column mapping and server-side behavior are documented in [Strava bulk import](../strava-bulk-import.md).
