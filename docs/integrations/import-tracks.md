# Import tracks (GPX / FIT)

Grom can import one or more GPX or FIT track files through the **system file picker**. This works in the **web** and **Android** clients. There is no Google Drive OAuth, Google Sign-In, or background folder sync.

## User flow

1. Open **Integration** → **External services**.
2. Under **Import tracks**, tap **Import tracks**.
3. In the system picker, select one or more `.gpx` / `.fit` files (from local storage, Downloads, or a cloud provider shown in the picker sidebar when available, such as Google Drive on Android).
4. Grom shows a progress bar while each file is parsed and uploaded, then a snackbar with how many workouts were created, skipped, invalid, or failed.

The Home app bar does not include a sync button for this flow.

## What gets imported

For each selected file:

| Source | Use in Grom |
|--------|-------------|
| Track parse (`POST /workouts/parse-track`) | Start time, durations, distance, speeds, optional activity name and sport type |
| Sport type | From the track when present (FIT sport/sub-sport or GPX track `<type>`); otherwise profile `last_sport_type`; otherwise `Run` |
| Equipment | Omitted on create so the server applies `last_equipment_by_sport` for the sport |
| Track blob | Attached on `POST /workouts` multipart |

Files that are not `.gpx`/`.fit` (by extension) are counted as invalid. Parse or create failures count as failed; remaining files continue.

## Duplicate detection (`external_id`)

| Field | Value |
|-------|--------|
| `external_id.name` | `device-import` |
| `external_id.id` | `{basename_lower}:{sha256_16}` — lowercase file name from the picker plus the first 16 hex characters of the SHA-256 of the file bytes |

`GET /workouts/external` skips files that already exist for the user. Renaming a file changes the id even if content is identical; changing content with the same name creates a new id.

Older workouts imported via the removed Health Sync + Google Drive flow keep their `health-sync/…` external ids; they are not migrated.

## Related

- [User overview](../user/overview.md)
- [Strava bulk import](strava-bulk-import.md) (server-side ZIP import)
- [Strava API import](strava-api-import.md) (Android BYO API sync)
