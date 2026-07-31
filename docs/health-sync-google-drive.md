# Health Sync + Google Drive import (Android)

Grom’s Android app can import workouts that [Health Sync](https://healthsync.app/) has exported to Google Drive. This flow is **Android-only**; the web UI does not expose it.

## User flow

1. In Health Sync, configure activity sync **to Google Drive**. Health Sync creates a Drive folder (typically named `Health Sync`) and writes per-activity CSV summaries plus track files (FIT preferred; GPX/TCX/KML when GPS is present).
2. In Grom, open **Integration**.
3. Enable **Health Sync + Google Drive sync**. Sign in with Google and grant read-only Drive access. Grom looks up a folder whose name contains `Health Sync` and fills **Health Sync folder**.
4. Optionally edit the folder name (exact match on sync) or tap the folder search icon to pick the first Drive folder matching `Health Sync*` again.
5. On **Home**, tap the sync icon in the header. Grom imports new workouts and refreshes the list.

## What gets imported

For each activity CSV in the folder:

| CSV column (by index) | Use in Grom |
|-----------------------|-------------|
| 1 Source app | Part of `external_id.name` as `health-sync/{source}` |
| 2 Sport | Mapped to a Grom sport type |
| 3 Title | Workout name |
| 4 Date/time | Start time (device local timezone; track values win when a track is attached) |
| 6 Elapsed | `duration_total_seconds` (track may override) |
| 7 Moving | `duration_seconds` (track may override) |
| 8 Distance (km) | Fallback if the track has no distance |

Track matching uses sport + date + time from the CSV filename: prefer `{date} {time}-{SPORT}.fit`, else `.gpx`. CSV-only activities (no track) are still created.

Already-imported rows are skipped using `external_id` (`name` + CSV filename as `id`).

## Google Cloud setup (APK builders)

Reading a user’s Drive requires an OAuth Android client in Google Cloud. This is for whoever **builds** the Android app, not for self-hosted server operators.

1. Create a Google Cloud project and enable **Google Drive API**.
2. Configure the **OAuth consent screen** (External + Testing is fine for development; add test users).
3. Add scope `https://www.googleapis.com/auth/drive.readonly`.
4. Create an **Android** OAuth client with:
   - Package name: `com.solargate.grom`
   - SHA-1 of the signing keystore (`keytool` on the debug keystore, or `./gradlew signingReport` under `ui/grom/android`)
5. Add a separate Android client for the **release** SHA-1 when publishing.

Google matches the app by package name + SHA-1; you usually do not embed the Android client ID in Flutter code. Misconfigured SHA-1/package typically surfaces as Google Sign-In error code 10.

## Related

- [User overview](user/overview.md)
- [Strava bulk import](strava-bulk-import.md) (server-side ZIP import; different from Health Sync)
