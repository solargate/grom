# Strava API import (Android)

On **Android**, Grom can import recent workouts directly from the [Strava API](https://developers.strava.com/) using **your own** Strava API application credentials (bring-your-own / BYO). Nothing is stored on the Grom server: Client ID, Client Secret, and OAuth tokens stay on the device. This flow is **not** available in the web client.

For a full history (or activities older than the sync window), use [Strava bulk import](strava-bulk-import.md) (ZIP archive).

## Prerequisites

1. A Strava account with an active subscription (required by Strava to create an API application).
2. Create an API application at [strava.com/settings/api](https://www.strava.com/settings/api).
3. Set **Authorization Callback Domain** to `localhost` (Grom uses redirect URI `grom://localhost/exchange_token`; a custom scheme is required so Android returns to the app after OAuth).
4. Copy the **Client ID** and **Client Secret** into Grom.

New Strava apps start in single-player mode (only the app owner can authorize), which matches personal BYO use.

## User flow

1. Open **Integration** → **External services** → **Strava** (Android only).
2. Enable **Import workouts from Strava**.
3. Enter Client ID and Client Secret.
4. Tap **Connect with Strava**, authorize with scope `activity:read`, and confirm the status shows connected.
5. On **Home**, tap the sync icon in the app bar (same placement/behavior as the former Health Sync button).
6. Grom shows a “Synchronizing…” dialog, then a snackbar with how many workouts were imported (or that none were new).

Turning the toggle **off** only hides the Home sync button; credentials and tokens remain on the device. Logging out of Grom also leaves Strava settings intact.

## Sync rules

| Rule | Behavior |
|------|----------|
| Limit | Up to **10** most recent activities (`kStravaApiSyncLimit`) |
| Order | Newest first |
| Stop | Stops at the first activity that already exists in Grom with `external_id.name=strava` and matching id |
| Visibility | `activity:read` only — Everyone / Followers activities (not “Only You”) |
| No GPS | Creates a workout from summary fields without a track |
| With GPS | Builds a GPX from Strava streams and attaches it |
| Photos | Best-effort download via the activity photos API; failures do not fail the workout |
| Equipment | `equipment_ids` omitted so the server applies `last_equipment_by_sport` |

Duplicate detection uses the same `external_id` namespace as the ZIP importer (`strava` + Strava activity id), so API and archive imports do not create duplicates for the same activity.

## Privacy note

Client secrets and refresh tokens are stored in app preferences on the device. Treat the phone like any other place you keep personal API credentials. Grom does not send these values to your Grom instance.

## Related

- [Strava bulk import](strava-bulk-import.md)
- [Import tracks](import-tracks.md)
- [User overview](../user/overview.md)
