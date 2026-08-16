# User overview

Grom’s Flutter client runs as a **web UI** and as an **Android** app. The web UI is served by the same `grom` process: open the server’s base URL in a browser (for example `http://localhost:8080/` with the default dev config). The screens and flows match the Android app; live GPS recording is Android-only. UI strings are available in English, Russian, and German.

When you are not signed in, **Home** shows a welcome screen with the Grom logo, a short description, and **Sign in** / **Register** buttons. On Android the welcome text also reminds you to enter the Grom server address when signing in.

On **Android** (and later iOS), sign-in and registration ask for a **server URL**. You can enter a bare host such as `grom.example.com` (no `https://` required). On submit the app probes `GET /api/v1/status` over HTTPS, then HTTP, writes the resolved URL into the field, and continues. If you already type `http://` / `https://` or an explicit port, that value is used as-is. **HTTP is supported for local / LAN instances** without TLS; prefer HTTPS for anything reachable on the public internet.

## Sign-in and password reset

If the operator enables outbound email (`mailer` in server config), the sign-in screen shows **Forgot password?**. Enter your account email; the server always responds the same way whether or not the address is registered. Check your inbox for a reset link and open it in a **browser** (the web UI at `/reset-password`). After you set a new password, sign in again in the app or on the web. Password reset is unavailable when the server reports `password_reset_enabled: false`.

When the operator enables captcha (`auth.captcha.enabled`), sign-in, registration, and forgot-password show an **I'm not a robot** checkbox. Tick it and wait for the local proof-of-work check to finish before submitting (it may take a moment on slower devices). Setting a new password from the email reset link does not require captcha.

## Profile and account deletion

From **Profile**, use the overflow menu to edit your profile or **Delete account**. Full steps, what data is removed, and federation notes: [Delete your Grom account](delete-account.md).

This page is a short tour of the main screens (screenshots below are from Android). Admin setup (install, config, TLS, federation) lives under [Admin docs](../README.md#admin). For the HTTP API, see Swagger at `/api/docs/` on a running server.

## Workouts

Your home feed lists activities with type, date, device, distance/time (and pace or elevation when relevant), plus a map preview when a GPS track is attached. Each card shows a social bar under the map and photos: likes on the left (thumb up to like or unlike someone else’s workout; tap the count for likers) and comments on the right (count + comment icon). You cannot like your own workouts; the like button stays disabled on yours. Anyone can comment on their own or others’ workouts. Tap the comment control to open the thread, add a comment (up to 1000 characters), or delete a comment you wrote (workout owners can also delete any comment on their workout).

![Workout list on Android](../screenshots/workout-list.jpg)

Open a workout for the full card: interactive map (when a track is present), photo gallery, the same social bar as in the list, a speed-over-distance chart with average and maximum speed, and a heart-rate chart (distance when GPS is present, otherwise elapsed minutes) with average and maximum heart rate. Tap a chart to see values at that point.

You can create workouts manually, import GPX/FIT tracks, attach photos, and link equipment. When editing a workout you can add or remove photos (up to 20). On create, the workout name defaults to the localized sport type and follows sport changes until you edit it. On Android you can also record a live GPS track (see below).

## Live recording (Android)

Open **Add workout → Record** to track a route on the map with live duration, speed, and distance. Keep the recording notification while the session is active. Auto-pause can be toggled on or off.

![Recording a workout on Android](../screenshots/workout-record.jpg)

## Equipment

Manage bikes, shoes, and other gear. Items are grouped by category; distance totals accumulate as you use gear on workouts (including after Strava bulk import).

![Equipment list on Android](../screenshots/equipment.jpg)

## Social and federation

Follow other users on the same instance and browse a shared feed. Like and comment on workouts from people you follow (local or federated); counts appear on list and detail screens.

When the operator enables ActivityPub federation, you can also follow athletes on other Grom instances (HTTPS required on the server). Likes on remote workouts are sent as ActivityPub `Like` activities; removing a like sends `Undo`. Comments on remote workouts are sent as `Create` Note with `inReplyTo`; deleting a comment sends `Delete`. Incoming likes and comments from other instances update the local workout the same way. Account deletion and federated actor `Delete` behavior are described in [Delete your Grom account](delete-account.md).

## Strava import

From **Integration → Strava**, upload a [Strava bulk data export](https://support.strava.com/hc/en-us/articles/216918437-Exporting-your-Data-and-Bulk-Export) ZIP. Grom imports activities, tracks, photos, and equipment. Column mapping and server-side behavior are documented in [Strava bulk import](../integrations/strava-bulk-import.md).

## Health Sync + Google Drive (Android)

On Android, **Integration** can enable Health Sync + Google Drive sync. Health Sync writes activity CSV/FIT/GPX files to Drive; Grom imports them from the Home sync button. Setup and OAuth notes: [Health Sync + Google Drive](../integrations/health-sync-google-drive.md).
