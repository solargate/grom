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
| Metriken | Distanz, Moving-/Elapsed-Zeit, Geschwindigkeiten, Höhe, Herzfrequenz, Kadenz, Leistung und Kalorien kommen aus dem Strava-Activity-Objekt, sofern vorhanden — nicht aus einer Neuberechnung des Tracks |
| Streams | Es werden alle Strava-Stream-Typen angefordert (`time`, `latlng`, `distance`, `altitude`, `velocity_smooth`, `heartrate`, `cadence`, `watts`, `temp`, `moving`, `grade_smooth`); fehlende Typen werden ignoriert |
| Track | Immer ein FIT aus Streams, wenn mindestens zwei nutzbare Samples existieren (Sensor und/oder GPS). Ohne gültiges GPS werden Positionsfelder weggelassen, damit Indoor-/Kraft-HF-Serien trotzdem anhängen. Leere oder durchgängig null-wertige Streams (z. B. Geschwindigkeit) werden nicht ins FIT geschrieben, damit unnötige Charts entfallen. Unvollständiges GPS ersetzt die Activity-Metriken nicht |
| Ohne Streams | Workout aus Summary ohne Track (avg/max HF, Kadenz, Watt, Kalorien werden trotzdem gesendet, wenn Strava sie liefert) |
| Gerät | Nutzt Strava `device_name`, falls vorhanden: wird in das erzeugte FIT (`product_name`) und als Create-Feld `device` geschrieben (sonst Server-Default `Grom App`). Das FIT setzt kein manufacturer=development, damit das Training nicht als „Development“ erscheint |
| Fotos | Best-effort über Photos-API; Foto-Fehler bricht das Workout nicht ab |
| Ausrüstung | Kein `equipment_ids` → Server nutzt `last_equipment_by_sport` |

Bereits importierte Aktivitäten werden vom Sync nicht aktualisiert (löschen und erneut importieren für einen korrigierten Track).

Dedup nutzt denselben `external_id`-Namespace wie der ZIP-Import.

## Datenschutz

Client Secret und Refresh-Tokens liegen in den App-Einstellungen auf dem Gerät. Grom sendet sie nicht an Ihre Instanz.

## Verwandt

- [Strava-Massenimport](strava-bulk-import.md)
- [Tracks importieren](import-tracks.md)
- [Benutzerüberblick](../user/overview.md)
