# Health Sync + Google Drive Import (Android)

Die Android-App von Grom kann Workouts importieren, die [Health Sync](https://healthsync.app/) nach Google Drive exportiert hat. Dieser Ablauf ist **nur unter Android**; die Web-UI stellt ihn nicht bereit.

## Nutzerablauf

1. Konfigurieren Sie in Health Sync die Aktivitätssynchronisation **nach Google Drive**. Health Sync legt einen Drive-Ordner an (typischerweise `Health Sync`) und schreibt CSV-Zusammenfassungen pro Aktivität plus Track-Dateien (FIT bevorzugt; GPX/TCX/KML bei GPS).
2. Öffnen Sie in Grom **Integration**.
3. Aktivieren Sie **Health Sync + Google Drive sync**. Melden Sie sich bei Google an und gewähren Sie schreibgeschützten Drive-Zugriff. Grom sucht einen Ordner, dessen Name `Health Sync` enthält, und füllt **Health Sync folder**.
4. Optional den Ordnernamen bearbeiten (exakter Match beim Sync) oder das Ordner-Such-Icon tippen, um erneut den ersten Drive-Ordner zu wählen, der auf `Health Sync*` passt.
5. Tippen Sie auf **Home** auf das Sync-Icon in der Kopfzeile. Grom importiert neue Workouts und aktualisiert die Liste.

Abmelden von Grom oder Ändern der Server-URL deaktiviert Health Sync und trennt den Google-Drive-Zugriff in der App. Aktivieren Sie den Schalter nach der Anmeldung erneut.

## Was importiert wird

Für jede Aktivitäts-CSV im Ordner:

| CSV-Spalte (nach Index) | Verwendung in Grom |
|-------------------------|--------------------|
| 1 Source app | Teil von `external_id.name` als `health-sync/{source}` |
| 2 Sport | Auf einen Grom-Sporttyp gemappt |
| 3 Title | Workout-Name |
| 4 Date/time | Startzeit (lokale Gerätezeitzone; Track-Werte gewinnen, wenn ein Track angehängt ist) |
| 6 Elapsed | `duration_total_seconds` (Track kann überschreiben) |
| 7 Moving | `duration_seconds` (Track kann überschreiben) |
| 8 Distance (km) | Fallback, wenn der Track keine Distanz hat |

Track-Matching nutzt Sport + Datum + Zeit aus dem CSV-Dateinamen: bevorzugt `{date} {time}-{SPORT}.fit`, sonst `.gpx`. Nur-CSV-Aktivitäten (ohne Track) werden trotzdem erzeugt.

Bereits importierte Zeilen werden über `external_id` übersprungen (`name` + CSV-Dateiname als `id`).

## Verwandt

- [Benutzerüberblick](../user/overview.md)
- [Strava-Massenimport](strava-bulk-import.md) (serverseitiger ZIP-Import; anders als Health Sync)
