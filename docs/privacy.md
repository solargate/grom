# Privacy Policy

**Last updated:** 2026-08-16

This Privacy Policy describes how the **Grom** Android application and web client (“App”) and the open-source Grom software handle information.

Grom is a **self-hosted** workout tracker. The App connects to a **Grom server URL that you choose**. Account data and workouts are stored on **that** server. There is no single central “official Grom cloud.” The App publisher may operate one or more public instances for testing or personal use; those are **not** designated as an official product cloud, and hosts may change over time.

- Documentation site (including this policy): [https://solargate.github.io/grom/](https://solargate.github.io/grom/)
- This policy on the web: [https://solargate.github.io/grom/privacy/](https://solargate.github.io/grom/privacy/)
- Source code: [https://github.com/solargate/grom](https://github.com/solargate/grom)
- Account deletion guide: [https://solargate.github.io/grom/user/delete-account/](https://solargate.github.io/grom/user/delete-account/)

**Contact**

- Name: Alexander Cheryomukhin  
- Email: [solargate.team@gmail.com](mailto:solargate.team@gmail.com)

If you use a third-party or self-hosted instance that you do not operate, the **operator of that server** decides retention, logging, email, captcha, federation, and related practices, and may publish their own notice. This policy explains what the App does and what a standard Grom deployment is designed to process.

## 1. Roles

| Role | Responsibility |
|------|----------------|
| **App publisher** (Play Store / this policy’s contact) | How the App accesses device data, local storage, and third-party services built into the client (for example map tiles and optional Google Drive). |
| **Server operator** | Account registration, workouts, media, social features, logs, and optional ActivityPub federation on the instance at the URL you configured. |
| **You** | Choosing a trustworthy server URL and reviewing that operator’s practices when the instance is not yours. |

If you connect to a Grom instance **operated by the App publisher**, the publisher processes personal data on that instance under this Privacy Policy (as server operator for that host). Connecting to any other URL means that other operator processes your account data.

## 2. Information the App accesses or processes

### 2.1 Account and authentication (sent to your chosen server)

When you register or sign in, the App may send to the configured server:

- Nickname
- Email address
- Password (transmitted for authentication; stored on the server as a password hash, not kept in clear text in the App)
- Session token (JWT) returned by the server

Optional, if the operator enables them: a captcha proof-of-work payload; password-reset email flow (reset links open in a browser).

### 2.2 Fitness and workout data (sent to your chosen server)

Depending on what you create, import, or record:

- Workout metadata (sport type, title, description, times, distances, device name, equipment links, notes)
- GPS tracks (precise location over time) from live recording, GPX/FIT import, Strava ZIP import (processed on the server), or Health Sync import
- Derived stats and charts (for example speed, elevation, cadence, power, heart rate, calories) when present in tracks or imports
- Photos you attach to workouts
- Avatar image
- Social data: follows, likes, and comments
- Personal access tokens (PAT) you create for API access
- Integration-related identifiers stored for imports (for example external workout IDs)

### 2.3 Precise and background location (device → App → your server)

On Android, live workout recording may use:

- Approximate and **precise** location
- **Background location** and a foreground location service, so recording can continue when you switch apps
- Notifications related to an active recording
- An optional battery-optimization exemption request so recording is less likely to be stopped by the system

Location is used **only** to record workout tracks and related map previews and stats. The App does not use location for advertising. Tracks you save are uploaded to **your configured Grom server**.

### 2.4 Photos and files

With your action, the App may read photos or GPX/FIT files you select (or that are shared into the App) and upload them to your configured server as workout media or tracks.

### 2.5 Data stored on the device

The App stores locally (for example via app preferences), including:

- Chosen server URL
- Session JWT
- UI preferences (for example locale and auto-pause)
- Optional Health Sync settings (for example Drive folder name and sync flags)

Clearing app data or signing out removes local session data according to the App’s logout behavior. Uninstalling the App does **not** delete your account on the server—use **Delete account** on that instance.

### 2.6 What the App publisher does not do via the App

The App does **not** include third-party advertising SDKs and does **not** include analytics or crash-reporting SDKs that report usage to the App publisher. The App publisher does **not** sell personal data.

Server operators may still configure their own logging, reverse proxies, mail providers, or backups outside the App client.

## 3. Third-party services used by the App or a Grom server

### 3.1 Your Grom server

All account and workout API traffic goes to the base URL you enter. Prefer HTTPS for any server reachable on the public internet. The operator of that host receives IP addresses and standard HTTP request metadata as part of normal server operation.

### 3.2 OpenStreetMap map tiles

When the App shows maps, it may request map tiles from OpenStreetMap tile servers. Those requests typically include your device IP address and standard HTTP headers. See OpenStreetMap’s privacy information and tile usage policy. Map display is for showing routes and the recording UI, not for selling location data.

### 3.3 Google Sign-In and Google Drive (optional, Android)

If you enable **Health Sync + Google Drive** import, the App uses Google Sign-In and requests **read-only** Google Drive access to import activity files that you (or Health Sync) placed in Drive. Google processes sign-in under Google’s policies. The App reads those files on the device and uploads imported workouts to **your configured Grom server**. The App publisher does not keep a separate copy of your Drive contents outside that chosen server. You can revoke Drive access in your Google Account settings.

### 3.4 ActivityPub federation (optional, server-side)

If the operator enables federation, workout and social content may be delivered to or received from other federated servers according to ActivityPub. Remote operators may retain copies even after local deletion if they never receive a Delete activity. See the [account deletion](user/delete-account.md) documentation.

### 3.5 Email (optional, server-side)

If the operator enables a mailer, password-reset messages are sent via that operator’s configured email transport (for example SMTP). The App publisher does not send those emails unless they operate the instance you use.

## 4. How information is used

| Purpose | Examples |
|---------|----------|
| Provide core features | Sign-in, workouts, tracks, photos, equipment, feed, likes, comments |
| Live recording | GPS track while you record on Android |
| Optional imports | Strava ZIP (on the server), Health Sync via Drive (on the device, then to the server) |
| Security | Auth tokens, optional captcha, rate limiting (may use IP on the server) |
| Federation | Optional delivery of activities to remote inboxes |

Data is used for App and server functionality you request—not for advertising by the App publisher.

## 5. Sharing

The App publisher does not sell your data.

Data may be disclosed or transferred when:

- You choose a server operator (they host your account)
- You use optional Google Drive import (Google)
- Map tiles are loaded (tile provider)
- Federation is enabled (remote ActivityPub servers)
- Required by law or to protect rights, safety, or integrity of the service
- You yourself export or share files (for example GPX)

## 6. Retention

- **On the device:** until you clear app data, sign out (session), or uninstall.
- **On a Grom server:** until you delete items or delete your account; operators may also retain backups or logs according to their own practices.
- **Federated copies:** may persist on remote servers beyond local deletion (Delete delivery is best-effort).

## 7. Security

Prefer a server with TLS (`https://`) for internet-facing instances. Passwords are handled as hashed credentials on the server. The App sends API traffic to your configured base URL. No security measure is perfect—self-hosted operators remain responsible for securing their deployment.

## 8. Your choices and account deletion

You can:

- Choose or change the server URL
- Deny or revoke OS permissions (location, notifications, photos); some features will not work without them
- Disable or disconnect Google Drive access
- Delete individual workouts or media where the product allows
- **Delete your account** in the App or web UI: Profile → Delete account (password required)

Details, including what is removed and federation notes: [Delete your Grom account](user/delete-account.md).

For questions about a publisher-operated instance, contact [solargate.team@gmail.com](mailto:solargate.team@gmail.com). For other instances, contact that server’s operator.

## 9. Children

The App is **not directed at children** and is not intended for Google Play Families / child-directed use. You must be at least **13** years old to create an account. We do not knowingly collect personal information from children under 13. If you believe a child under 13 has provided personal data, contact us and we will take reasonable steps to delete it from instances we operate.

## 10. International processing

Because you may connect to a server in another country, your data may be processed in a country different from where you live. Choose an operator and region you trust. This policy describes practices globally in clear language; applicable privacy laws where you use the App or where a server is operated may also apply.

## 11. Additional information for EEA and UK users

If you are in the European Economic Area or the United Kingdom, you may have rights under applicable law regarding access, rectification, erasure, restriction, objection, and data portability, and the right to lodge a complaint with a supervisory authority. To exercise rights regarding an instance operated by the App publisher, contact [solargate.team@gmail.com](mailto:solargate.team@gmail.com). For other instances, contact that operator.

Processing on a publisher-operated instance is typically necessary to provide the service you request (account and workouts), and optional features such as background location or Google Drive import rely on your consent or affirmative use of those features.

## 12. Changes

We may update this policy. The “Last updated” date will change, and the current version will be published at [https://solargate.github.io/grom/privacy/](https://solargate.github.io/grom/privacy/). Where required by law, we will provide additional notice of material changes.

## 13. Open-source software

Grom is released under the GPL-3.0 license. This Privacy Policy concerns product data practices, not the software license.
