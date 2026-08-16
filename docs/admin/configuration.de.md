# Konfiguration

Grom wird über eine YAML-Datei konfiguriert. Standardmäßig sucht er `config.yaml` im aktuellen Arbeitsverzeichnis. Mit `--config` (oder `-c`) wählen Sie einen anderen Pfad.

- Beispielprofile: [`cmd/grom/config-examples/`](https://github.com/solargate/grom/tree/master/cmd/grom/config-examples)
- Alle Felder mit Kommentaren: [`config.full.yaml`](https://github.com/solargate/grom/blob/master/cmd/grom/config-examples/config.full.yaml)

**Erforderlich:** `auth.jwt_secret` — ein langes Zufallsgeheimnis zum Signieren von JWT-Access-Token.

## Wichtige Einstellungen

| Bereich | Was setzen |
|---------|------------|
| `server.port` / `server.tls` | Listen-Ports und TLS-Modus (`off`, `static` oder `autocert`) |
| `storage.driver` / `location` / `temp_dir` | `file` (Standard; Tests / winzige Instanzen) oder `bbolt` (empfohlen für normale Installationen); Datenwurzel und Temp-Verzeichnisse |
| `storage.bbolt.path` | Optionaler Pfad zu `grom.db` bei bbolt (Standard: `{location}/grom.db`) |
| `federation.enabled` / `federation.domain` | ActivityPub; erfordert HTTPS |
| `auth.reset` / `mailer` | Passwort-Reset per E-Mail (`public_base_url`, SMTP oder Log-Treiber) |
| `auth.captcha` | Optionales ALTCHA-PoW bei Registrierung/Login/„vergessen“ (`enabled`, optional `hmac_secret` / `cost` / `expires_seconds`) |
| `logging.level` / `logging.format` | `debug`/`info`/`warn`/`error`; `text` (Dev) oder `json` (Prod). Standard: `info` + `json`. Gin-Framework-Debug (`[GIN-debug]`) nur bei `logging.level: debug`; sonst Gin im Release-Modus |

Relative Pfade in `storage.*`, `server.tls.cert_file` / `key_file`, `server.tls.autocert.cache_dir` und `federation.ca_cert_file` werden relativ zum Verzeichnis der `grom`-Binary aufgelöst (absolute Pfade bleiben unverändert).

## TLS-Profile

| Profil | Config-Datei | `tls.mode` | Föderation |
|--------|--------------|------------|------------|
| Dev, nur HTTP | `config.dev.notls.yaml` | `off` | deaktiviert |
| Dev, self-signed TLS | `config.dev.tls.yaml` | `static` | aktiviert |
| Prod, nur HTTP | `config.prod.notls.yaml` | `off` | deaktiviert |
| Prod, Let's Encrypt | `config.prod.tls.yaml` | `autocert` | aktiviert |

**Dev mit static TLS** — Zertifikate erzeugen, dann starten:

```bash
cd cmd/grom
go run . gencerts --ip 192.168.1.251 --domain 192.168.1.251
go run . --config config-examples/config.dev.tls.yaml
```

Für Föderation zwischen lokalen Instanzen setzen Sie `federation.tls_insecure_skip_verify: true` und optional `federation.ca_cert_file`, um Ihrer Dev-CA zu vertrauen.

**Production mit autocert** — öffentlicher DNS-Name in `federation.domain` (nur Hostname), Ports **80** und **443** aus dem Internet erreichbar. ACME-Zertifikate werden standardmäßig unter `acme-cache` neben der grom-Binary gecacht (Override: `server.tls.autocert.cache_dir`; absoluten Pfad nutzen, wenn die Binary unter einem Systemverzeichnis wie `/usr/bin` liegt):

```bash
go run . --config config-examples/config.prod.tls.yaml
```

Hinweise:

- Föderation erfordert HTTPS (`tls.mode: static` oder `autocert`). Mit `tls.mode: off` läuft sie nicht.
- Bei aktivierter Föderation werden Likes auf Remote-Workouts als ActivityPub `Like` / `Undo` zugestellt, Kommentare als `Create`/`Delete` Note (`inReplyTo`); lokale Workouts akzeptieren eingehende Likes und Kommentare von anderen Instanzen. Likes und Kommentare auf derselben Instanz funktionieren ohne Föderation. Kontolöschung (`DELETE /api/v1/auth/me`) liefert ein `Delete` des lokalen Actors an bekannte Remote-Inboxes (best-effort), bevor lokale Daten gelöscht werden; die Inbox wendet Remote-Actor-`Delete` an, indem der föderierte Cache dieses Besitzers für den Empfänger bereinigt wird.
- Legacy-Configs mit `server.tls.enabled: true` (ohne `mode`) gelten als `mode: static`.

## Speicher-Treiber

`file` eignet sich für Tests und sehr kleine Instanzen (YAML-Metadaten auf der Platte sind leicht einsehbar). Für normale oder Production-Installationen bevorzugen Sie `bbolt` — Metadaten in einer Bolt-DB, Tracks, Fotos und andere Blobs bleiben auf dem Dateisystem.

| Treiber | Metadaten | Diagramme (Geschwindigkeit / Herzfrequenz) | Workout-Likes / -Kommentare | Blobs (Tracks, Fotos, Avatare, Schlüssel) |
|---------|-----------|--------------------------------------------|-----------------------------|-------------------------------------------|
| `file` (Standard) | YAML unter `storage.location` (`users.yaml`, pro Nutzer `equipment.yaml`, `profile.yaml`, Workout-YAML, …) | `speed-chart.json` und `heartrate-chart.json` im jeweiligen Workout-Verzeichnis | `likes.yaml` und `comments.yaml` neben jedem lokalen Workout; föderierter Like-/Kommentar-Cache und Outbox-Activity-IDs unter dem `federation/`-Baum des Betrachters | Dieselbe Baumstruktur |
| `bbolt` | JSON in `{location}/grom.db` (oder `storage.bbolt.path`); inkl. Bucket `user_profiles` für UI-Präferenzen | Gepackte Binärwerte in bbolt-Buckets `speed_charts` / `fed_speed_charts` und `heart_rate_charts` / `fed_heart_rate_charts` (föderierte Inbox) | Buckets `workout_likes`, `fed_workout_likes`, `like_activities`, `workout_comments`, `fed_workout_comments` und `comment_activities` | Dasselbe Dateisystem-Layout unter `storage.location` |

`postgres` ist in der Config reserviert, aber nicht implementiert.

Metadaten zwischen Treibern migrieren (Server zuerst stoppen; Track-/Medien-/Avatar-Blobs sind gemeinsam und werden nicht kopiert). Geschwindigkeits- und Herzfrequenz-Diagramme werden zwischen Datei-JSON-Blobs und bbolt-Binärbuckets konvertiert. Workout-Likes und -Kommentare (lokal, föderierter Cache und ausgehende Activity-IDs) sowie persönliche Zugriffstoken werden kopiert, damit sie nach dem Treiberwechsel lesbar bleiben:

| Kopiert von `migrate-storage` | Nicht kopiert |
|-------------------------------|---------------|
| Nutzer, Profile, Ausrüstung, Workouts | Passwort-Reset-Token (kurzlebig; laufende Reset-Links werden ungültig) |
| Follows, Föderations-Follower, föderierte Inbox-Autoren/Workouts | Blob-Dateien (Tracks, Fotos, Avatare, Schlüssel) — gemeinsam auf der Platte |
| Lokale/föderierte Likes und Kommentare + ausgehende Activity-IDs | Temporäre Strava-Jobs / Captcha-Zustand im Speicher |
| Speed-/HR-Diagramme (Formatkonvertierung) | |
| Persönliche Zugriffstoken (PAT) | |

```bash
grom migrate-storage --config config.yaml --from file --to bbolt --verify
# dann storage.driver: bbolt setzen und neu starten
grom migrate-storage --config config.yaml --from bbolt --to file --verify
```

`--dry-run` zählt Datensätze ohne Schreiben; `--force` überschreibt eine vorhandene bbolt-Datenbank.

Passwort-Reset-Token (`reset_tokens.yaml` / bbolt `reset_tokens`) sind kurzlebig und werden von migrate-storage **nicht** kopiert; laufende Reset-Links werden nach einer Migration ungültig. Legacy-Plaintext-Like-Activity-IDs (ohne `object_id`) werden über die föderierte Inbox und, wenn `federation.domain` gesetzt ist, über lokale Workout-Objekt-URLs rekonstruiert.

## Mailer und Passwort-Reset

Ausgehende E-Mail ist optional. Bei `mailer.driver: off` (Standard) ist Passwort-Reset deaktiviert und `GET /api/v1/server-info` meldet `password_reset_enabled: false`.

| Einstellung | Zweck |
|-------------|-------|
| `auth.reset.public_base_url` | Basis-URL in Reset-Links (ohne trailing slash). Erforderlich, wenn der Mailer an ist. |
| `auth.reset.token_ttl_minutes` | Token-Lebensdauer (Standard `60`) |
| `mailer.driver` | `off`, `log` (in den Server-Log schreiben — nützlich in Dev) oder `smtp` |
| `mailer.from` | Absenderadresse |
| `mailer.smtp.host` / `port` | SMTP-Relay (erforderlich bei `driver: smtp`). Übliche Ports: `587` (STARTTLS), `465` (implizites TLS) |
| `mailer.smtp.username` / `password` | Optionale SMTP-Credentials |
| `mailer.smtp.encryption` | `starttls` (Standard; auch Standard für Port 587), `tls` (Standard bei Port `465`) oder `none` |

Es gibt **keine** Abhängigkeit von lokalem MTA / `sendmail`: der Prozess spricht SMTP (über [go-mail](https://github.com/wneessen/go-mail)) mit einem externen Anbieter (Gmail-App-Passwort, SES, Mailgun usw.) oder loggt die Nachricht bei `driver: log`.

Passwort-Reset-Endpunkte nutzen einen In-Memory-Rate-Limiter mit festem Fenster (15 Minuten): forgot — 10 Anfragen pro Client-IP und 3 pro E-Mail; Confirm-Reset — 20 pro Client-IP. Limits nutzen Gins `ClientIP()` (berücksichtigt `X-Forwarded-For` / `X-Real-IP`, falls vorhanden). Grom exponiert noch keine Trusted-Proxies-Einstellung; behandeln Sie Forwarded-Header als untrusted, solange Ihr Reverse Proxy sie nicht überschreibt oder entfernt.

Beispiel (Production-SMTP auf Port 587):

```yaml
auth:
  jwt_secret: "..."
  reset:
    public_base_url: "https://grom.example.com"
mailer:
  driver: smtp
  from: "Grom <noreply@grom.example.com>"
  smtp:
    host: smtp.example.com
    port: 587
    username: "apikey"
    password: "secret"
    encryption: starttls
```

## Auth-Captcha (ALTCHA)

Optionales self-hosted [ALTCHA](https://altcha.org/)-Proof-of-Work-Captcha (kein Drittanbieterdienst, keine API-Keys). Bei `auth.captcha.enabled: true`:

- **Geschützt:** `POST /api/v1/auth/register`, `/auth/login` und `/auth/password/forgot` — Request-Body muss ein gelöstes `altcha`-Payload enthalten (base64-JSON).
- **Nicht geschützt:** `POST /api/v1/auth/password/reset` (Token aus dem E-Mail-Link reicht).
- **Challenge:** `GET /api/v1/captcha/challenge` liefert eine PoW-Challenge (`200`). Bei ausgeschaltetem Captcha → `404`; bei Challenge-Rate-Limit für die IP → `429` mit `Retry-After`.
- **Client-Erkennung:** `GET /api/v1/server-info` enthält `captcha_enabled`. Die Flutter-Web-/Android-UI zeigt eine Checkbox **I'm not a robot** / „Ich bin kein Roboter“ und löst die Challenge lokal vor dem Absenden.

| Einstellung | Zweck |
|-------------|-------|
| `auth.captcha.enabled` | Captcha bei Registrierung/Login/„vergessen“ verlangen. Standard: `false`. |
| `auth.captcha.hmac_secret` | HMAC-Schlüssel für Challenge-Signaturen. Optional; wenn leer, wird `auth.jwt_secret` genutzt. In Production lieber ein separates Geheimnis, wenn Sie JWT unabhängig rotieren. |
| `auth.captcha.cost` | PBKDF2-Iterationskosten für PoW. Standard: `1000`. Höhere Werte belasten die Client-CPU stärker. |
| `auth.captcha.expires_seconds` | Challenge-Lebensdauer in Sekunden. Standard: `300`. |

Challenge-Ausgabe ist In-Memory rate-limited (60 Anfragen pro Client-IP pro 15-Minuten-Fenster). Gelöste Payloads sind einmalig bis zum Ablauf (Replay abgelehnt). Rate Limits und Replay-Store sind prozesslokal — sie resetten beim Neustart und werden nicht über mehrere Grom-Prozesse geteilt; bei Multi-Instance TLS/Rate-Limit an einem Reverse Proxy terminieren oder Captcha aus lassen, bis Sticky Sessions / gemeinsamer Store existieren.

Client-IP-Auflösung wie beim Passwort-Reset: Gins `ClientIP()` (berücksichtigt `X-Forwarded-For` / `X-Real-IP`). Ohne vertrauenswürdigen Reverse Proxy, der diese Header überschreibt, behandeln Sie sie als untrusted.

Verifikationsfehler liefern `400` mit Meldungen wie `captcha is required`, `invalid captcha`, `captcha expired` oder `captcha already used`.

Beispiel:

```yaml
auth:
  jwt_secret: "..."
  captcha:
    enabled: true
    # hmac_secret: "optional-separate-secret"
    cost: 1000
    expires_seconds: 300
```

## Siehe auch

- [Installation und Start](install.md)
- [Strava-Massenimport](../integrations/strava-bulk-import.md) (`storage.temp_dir`)
