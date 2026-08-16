# Grom-API-Token

Persönliche Zugriffstoken (PAT) ermöglichen den Zugriff auf die Grom-HTTP-API aus Skripten und externen Apps, ohne die Web- oder Mobile-Login-Sitzung zu nutzen.

## Token verwalten

Öffnen Sie **Integration** → **Grom** in der Web- oder Mobile-App. Sie können bis zu **10** Token pro Konto erstellen und jederzeit widerrufen.

Beim Erstellen zeigt Grom das vollständige Geheimnis **einmal**. Kopieren Sie es sofort und speichern Sie es sicher (Passwortmanager, OS-Keychain oder Secret-Store Ihrer Automatisierung). Das Geheimnis lässt sich später nicht erneut anzeigen; die Liste zeigt nur ein kurzes Präfix (z. B. `grom_pat_a1…`).

## Token-Format

```
Authorization: Bearer grom_pat_<secret>
```

Die Standardlebensdauer beträgt **90 Tage**. Sie können eine andere Dauer wählen oder explizit ein Token **ohne Ablauf** erstellen (mit Warnung in der UI).

## Scopes

| Scope | Erlaubt |
|-------|---------|
| `workouts:read` | Eigene lokale Workouts, Tracks, Diagramme, Kartenvorschauen und Medien auflisten und lesen |
| `workouts:write` | Workouts erstellen, aktualisieren, löschen; Tracks und Medien hochladen |
| `equipment:read` | Ausrüstung auflisten |
| `equipment:write` | Ausrüstung erstellen, aktualisieren, löschen |

Likes, Kommentare, Social-Features, Strava-Import und Profil-Endpunkte erfordern einen normalen UI-Login (JWT), kein PAT.

PAT-Workout-Zugriff ist auf **Ihre eigenen lokalen Workouts** beschränkt (nicht der Social Feed und nicht föderierte Workouts anderer Instanzen).

## Beispiel

```bash
# Ihre Workouts auflisten (PAT mit workouts:read)
curl -sS -H "Authorization: Bearer grom_pat_YOUR_TOKEN" \
  "https://your-grom.example/api/v1/workouts?scope=own"

# Token erstellen (JWT vom Login nötig — für den Alltag besser die App-UI)
curl -sS -X POST -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{"name":"Backup","scopes":["workouts:read"],"expires_in_days":90}' \
  "https://your-grom.example/api/v1/auth/pat"
```

Vollständige API-Details: Swagger UI unter `/api/docs/` auf Ihrem Server.

## Speicherung und Migration

Token werden auf dem Server in `personal_access_tokens.yaml` (File-Treiber) oder im bbolt-Bucket `personal_access_tokens` gespeichert. `grom migrate-storage` kopiert Token zwischen Treibern, damit bestehende `grom_pat_…`-Geheimnisse nach dem Wechsel von `storage.driver` weiter funktionieren.
