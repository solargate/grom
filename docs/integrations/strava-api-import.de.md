# Strava-API-Import (Android)

Unter **Android** kann Grom aktuelle Trainings direkt über die [Strava API](https://developers.strava.com/) importieren – mit **Ihren eigenen** Strava-API-Anwendungsdaten (Bring-your-own / BYO). Auf dem Grom-Server wird nichts gespeichert: Client ID, Client Secret und OAuth-Tokens bleiben auf dem Gerät. Im Web-Client ist dieser Ablauf **nicht** verfügbar.

Für die vollständige Historie (oder ältere Aktivitäten) nutzen Sie den [Strava-Massenimport](strava-bulk-import.md) (ZIP-Archiv).

## Voraussetzungen

1. Strava-Konto mit aktivem Abo (von Strava für API-Apps verlangt).
2. API-Anwendung unter [strava.com/settings/api](https://www.strava.com/settings/api) anlegen.
3. **Client ID** und **Client Secret** in Grom eintragen.

Wenn Connect mit HTTP **403** beim Token-Austausch scheitert, prüfen Sie Client ID und Client Secret und versuchen Sie es erneut.

Neue Strava-Apps starten im Single-Player-Modus (nur der App-Besitzer kann autorisieren) – passend für persönlichen BYO-Einsatz.

## Ablauf

1. **Integration** → **Externe Dienste** → **Strava** öffnen (nur Android).
2. **Trainings aus Strava importieren** aktivieren.
3. Client ID, Client Secret und optional **Trainings pro Sync** eingeben (Standard **10**, max. **200**).
4. **Connect with Strava** tippen, Scope `activity:read` erteilen, Status **OK** prüfen.
5. Auf **Start** das Sync-Symbol in der App-Leiste tippen (wie früher bei Health Sync).
6. Dialog „Synchronisiere…“, danach Snackbar mit Importanzahl (oder keine neuen Trainings).

Umschalter aus = Sync-Button ausgeblendet; Credentials/Tokens bleiben. Logout aus Grom löscht sie ebenfalls nicht.

## Sync-Regeln

| Regel | Verhalten |
|-------|-----------|
| Limit | Bis zur konfigurierbaren Anzahl **Trainings pro Sync** (Standard **10**, max. **200**) der neuesten Aktivitäten |
| Reihenfolge | Neueste zuerst |
| Stopp | Bei der ersten bereits vorhandenen Aktivität mit `external_id.name=strava` |
| Sichtbarkeit | Nur `activity:read` — Everyone / Followers (nicht „Only You“) |
| Ohne GPS | Workout aus Summary ohne Track |
| Mit GPS | GPX aus Strava-Streams |
| Gerät | Nutzt Strava `device_name`, falls vorhanden (sonst Server-Default `Grom App`) |
| Fotos | Best-effort über Photos-API; Foto-Fehler bricht das Workout nicht ab |
| Ausrüstung | Kein `equipment_ids` → Server nutzt `last_equipment_by_sport` |

Dedup nutzt denselben `external_id`-Namespace wie der ZIP-Import.

## Datenschutz

Client Secret und Refresh-Tokens liegen in den App-Einstellungen auf dem Gerät. Grom sendet sie nicht an Ihre Instanz.

## Verwandt

- [Strava-Massenimport](strava-bulk-import.md)
- [Tracks importieren](import-tracks.md)
- [Benutzerüberblick](../user/overview.md)
