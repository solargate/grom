# Dokumentation

Grom ist ein self-hosted Trainings-Tracker mit optionaler ActivityPub-Föderation. Erfassen oder importieren Sie Workouts, verwalten Sie Ausrüstung, folgen Sie anderen Sportlern (lokal oder föderiert), liken und kommentieren Sie Aktivitäten und browsen Sie einen Social Feed — auf Infrastruktur, die Sie kontrollieren.

Der Server ist eine einzelne Go-Binary; der Flutter-Client erscheint als eingebettete Web-UI und als Android-App.

## Funktionen

- **Workouts** — Aktivitäten mit Stats, Notizen, Medien und Kartenvorschauen erstellen und bearbeiten
- **GPS-Tracks** — GPX und FIT importieren; Live-Aufzeichnung auf Android
- **Ausrüstung** — Fahrräder, Schuhe und anderes Gear, verknüpft mit Workouts
- **Social Feed** — Nutzern folgen und ihre Workouts in einer Timeline sehen
- **Likes** — fremde Aktivitäten liken, Zähler und Likes-Liste; föderierte `Like` / `Undo` bei aktiviertem ActivityPub
- **Kommentare** — eigene und fremde Aktivitäten kommentieren (hinzufügen/auflisten/löschen); föderierte `Create`/`Delete` Note bei aktiviertem ActivityPub
- **Föderation** — optionales ActivityPub, damit Instanzen sich netzwerkweit folgen können
- **Strava-Import** — Massenimport eines Strava-Datenexport-ZIP
- **Health Sync** — Aktivitäten von Google Drive auf Android importieren (Health-Sync-Exporte)
- **API-Token** — persönliche Zugriffstoken mit Scopes (`grom_pat_…`) für Workouts und Ausrüstung
- **Clients** — dieselbe Flutter-UI im Browser (vom Server ausgeliefert) und als Android-APK
- **Sprachen** — Englisch, Russisch und Deutsch in der Flutter-UI

<p align="center">
  <img src="screenshots/workout-list.jpg" width="250" alt="Workout-Liste" />
  <img src="screenshots/workout-record.jpg" width="250" alt="Live-Aufzeichnung" />
  <img src="screenshots/equipment.jpg" width="250" alt="Ausrüstung" />
</p>

Tour durch die Bildschirme: [Benutzerüberblick](user/overview.md). Bauen und starten: [Installation und Start](admin/install.md). Quellcode und Quick Start: [Repository-README](https://github.com/solargate/grom#readme).

## Ich möchte…

### Benutzer

| Ziel | Seite |
|------|-------|
| Sehen, was der Client kann (Workouts, Likes, Kommentare, Aufzeichnung, Ausrüstung) | [Benutzerüberblick](user/overview.md) |
| Ein vergessenes Passwort zurücksetzen (wenn der Betreiber E-Mail aktiviert) | [Überblick — Anmeldung und Passwort-Reset](user/overview.md#anmeldung-und-passwort-reset) |
| Anmelden / registrieren, wenn die Instanz Captcha aktiviert | [Überblick — Anmeldung und Passwort-Reset](user/overview.md#anmeldung-und-passwort-reset) |
| Grom im Browser nutzen (gleiche UI wie Android) | Basis-URL des Servers nach der [Installation](admin/install.md) öffnen; siehe [Benutzerüberblick](user/overview.md) |
| Einen Strava-Export importieren (UI + Importverhalten) | [Strava-Massenimport](integrations/strava-bulk-import.md) |
| Health-Sync-Aktivitäten von Google Drive importieren (Android) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| API-Token für Skripte und externe Apps erstellen | [Grom-API-Token](user/grom-api-tokens.md) |
| Mein Konto löschen | [Grom-Konto löschen](user/delete-account.md) |

### Administration

| Ziel | Seite |
|------|-------|
| Server bauen und starten | [Installation und Start](admin/install.md) |
| TLS, Speicher, Föderation, Logging, Mailer / Passwort-Reset, Captcha konfigurieren | [Konfiguration](admin/configuration.md) |

### Referenz

| Thema | Seite |
|-------|-------|
| Strava-ZIP-Spaltenzuordnung und Importverhalten | [Strava-Massenimport](integrations/strava-bulk-import.md) |
| Health-Sync-Import von Google Drive (Android-Client) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Vollständige annotierte Konfiguration | [`config.full.yaml`](https://github.com/solargate/grom/blob/master/cmd/grom/config-examples/config.full.yaml) |
| HTTP-API (Swagger UI) | Auf dem laufenden Server: `/api/docs/` (z. B. `http://localhost:8080/api/docs/`); OpenAPI-Quellen unter [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) |
| Datenschutzrichtlinie | [Privacy Policy](privacy.md) |
