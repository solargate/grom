# Grom API tokens

Personal access tokens (PAT) let you access the Grom HTTP API from scripts and external apps without using the web or mobile login session.

## Where to manage tokens

Open **Integration** → **Grom** in the web or mobile app. You can create up to **10** tokens per account and revoke them at any time.

When you create a token, Grom shows the full secret **once**. Copy it immediately and store it securely (password manager, OS keychain, or your automation platform’s secret store). You cannot view the secret again later; the list only shows a short prefix (for example `grom_pat_a1…`).

## Token format

```
Authorization: Bearer grom_pat_<secret>
```

Default lifetime is **90 days**. You can choose another duration or explicitly create a token **without expiration** (with a warning in the UI).

## Scopes

| Scope | Allows |
|-------|--------|
| `workouts:read` | List and read your own local workouts, tracks, charts, map previews, and media |
| `workouts:write` | Create, update, delete workouts; upload tracks and media |
| `equipment:read` | List equipment |
| `equipment:write` | Create, update, delete equipment |

Likes, comments, social features, Strava import, and profile endpoints require a normal UI login (JWT), not a PAT.

PAT workout access is limited to **your own local workouts** (not the social feed and not federated workouts from other instances).

## Example

```bash
# List your workouts (PAT with workouts:read)
curl -sS -H "Authorization: Bearer grom_pat_YOUR_TOKEN" \
  "https://your-grom.example/api/v1/workouts?scope=own"

# Create a token (requires JWT from login — use the app UI instead for day-to-day use)
curl -sS -X POST -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{"name":"Backup","scopes":["workouts:read"],"expires_in_days":90}' \
  "https://your-grom.example/api/v1/auth/pat"
```

Full API details: Swagger UI at `/api/docs/` on your server.

## Storage and migration

Tokens are stored on the server in `personal_access_tokens.yaml` (file driver) or the `personal_access_tokens` bbolt bucket. They are **not** copied by `grom migrate-storage`; recreate tokens after migrating storage drivers.
