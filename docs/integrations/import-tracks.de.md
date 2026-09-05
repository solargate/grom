# Tracks importieren (GPX / FIT)

Grom kann eine oder mehrere GPX- oder FIT-Trackdateien über den **System-Dateidialog** importieren. Das funktioniert in den Clients **Web** und **Android**. Es gibt kein Google-Drive-OAuth, kein Google Sign-In und keine Hintergrund-Synchronisation eines Ordners.

## Ablauf

1. Öffnen Sie **Integration** → **External services**.
2. Unter **Tracks importieren** tippen Sie auf **Tracks importieren**.
3. Wählen Sie im Systemdialog eine oder mehrere `.gpx`-/`.fit`-Dateien (lokal, Downloads oder ein Cloud-Anbieter in der Seitenleiste des Pickers, falls verfügbar, z. B. Google Drive unter Android).
4. Während des Parsens und Hochladens erscheint ein Fortschrittsbalken; danach ein Snackbar mit erstellten, übersprungenen, ungültigen und fehlgeschlagenen Einträgen.

In der Home-App-Leiste gibt es dafür keinen Sync-Button.

## Was importiert wird

Pro ausgewählter Datei:

| Quelle | Verwendung in Grom |
|--------|--------------------|
| Track-Parse (`POST /workouts/parse-track`) | Startzeit, Dauern, Distanz, Geschwindigkeiten, optional Name und Sportart |
| Sportart | Aus dem Track, wenn vorhanden (FIT Sport/SubSport oder GPX `<type>`); sonst Profil `last_sport_type`; sonst `Run` |
| Equipment | Beim Erstellen weggelassen — der Server nutzt `last_equipment_by_sport` |
| Track | Als Multipart an `POST /workouts` |

Dateien ohne Endung `.gpx`/`.fit` gelten als ungültig. Parse-/Create-Fehler zählen als failed; die übrigen Dateien werden weiter verarbeitet.

## Duplikaterkennung (`external_id`)

| Feld | Wert |
|------|------|
| `external_id.name` | `device-import` |
| `external_id.id` | `{basename_lower}:{sha256_16}` — Dateiname in Kleinbuchstaben plus die ersten 16 Hex-Zeichen des SHA-256 der Bytes |

`GET /workouts/external` überspringt bereits vorhandene Paare. Umbenennen ändert die id bei gleichem Inhalt; geänderter Inhalt bei gleichem Namen erzeugt eine neue id.

Ältere Workouts aus Health Sync + Google Drive behalten `health-sync/…`; es gibt keine Migration.

## Verwandt

- [Benutzerübersicht](../user/overview.md)
- [Strava-Massenimport](strava-bulk-import.md) (serverseitiger ZIP-Import)
- [Strava-API-Import](strava-api-import.md) (Android BYO-API-Sync)
