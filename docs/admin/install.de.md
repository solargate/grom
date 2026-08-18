# Installation und Start

Installieren Sie Grom aus einem GitHub-Release oder bauen Sie aus dem Quellcode. Konfigurationsdetails (TLS, Speicher, Föderation, Mailer / Passwort-Reset, Captcha) siehe [Konfiguration](configuration.md).

## Download von GitHub Releases

Vorgefertigte Pakete erscheinen mit jedem [GitHub-Release](https://github.com/solargate/grom/releases):

| Asset | Inhalt |
|-------|--------|
| `grom-<version>-linux-amd64.tar.gz` | Server-Binary + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Server-Binary + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Server-Binary + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Server-Binary + `config-examples/` |
| `grom-<version>.apk` | Android-Client |

Entpacken Sie das Archiv für Ihr OS, kopieren Sie eine Konfiguration aus `config-examples/` (setzen Sie `auth.jwt_secret`) und starten Sie die `grom`-Binary — siehe [Start](#start) unten.

## Aus dem Quellcode bauen

### Voraussetzungen

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (für die eingebettete Web-UI und Android-Builds)
- Python 3 (MkDocs-Dokumentationssite und `server-catalog.yaml`-Tooling; `make web` / `make android-*` erzeugen den Katalog neu)
- Make

### Build

Aus dem Repository-Root:

```bash
make grom    # swagger + Flutter web + Go-Binary → cmd/grom/grom
make test    # Go- und Flutter-Tests
make cli     # nur Go-Binary (ohne Flutter-Web-Rebuild)
make web     # Flutter web → internal/web/dist
```

Weitere nützliche Targets: `make apidoc` (OpenAPI neu erzeugen), `make catalog` (Server-Katalog-Dart), `make android-apk`, `make clean`.

## Start

Beispielkonfigurationen liegen in den Release-Archiven und unter `cmd/grom/config-examples/` im Repo. Der Server sucht `config.yaml` im aktuellen Arbeitsverzeichnis, sofern Sie nicht `--config` / `-c` übergeben.

**Aus einem Release-Paket** (nur-HTTP-Beispiel):

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# config.yaml bearbeiten — auth.jwt_secret setzen
./grom --config config.yaml
```

**Aus dem Quellcode, nur HTTP** (kein TLS, Föderation aus):

```bash
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
```

**Nach `make grom`**, mit Config neben der Binary:

```bash
cd cmd/grom
./grom --config config.yaml
# oder Standard-config.yaml im Arbeitsverzeichnis:
./grom
```

Öffnen Sie die **Web-UI** im Browser unter der Basis-URL des Servers (derselbe Flutter-Client wie Android — z. B. `http://localhost:8080/` mit `config.dev.notls.yaml`). Registrieren Sie einen Nutzer und melden Sie sich an. Siehe [Benutzerüberblick](../user/overview.md).

Die **Android**-App (später iOS) kann sich mit derselben Instanz verbinden: Host auf dem Login-Bildschirm eingeben (Schema optional) oder einen [freigegebenen öffentlichen Server](../user/approved-servers.md) aus der Liste wählen. Klartext-**HTTP ist für lokale/LAN**-Installationen ohne TLS erlaubt; nutzen Sie HTTPS, wenn der Server über das lokale Netz hinaus erreichbar ist. Wie der Client `http` vs `https` auflöst: [Benutzerüberblick](../user/overview.md).

**API-Dokumentation (Swagger UI):** `http://<host>:<port>/api/docs/` (z. B. `http://localhost:8080/api/docs/`). Generierte OpenAPI-Quellen liegen auch unter [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) im Repository.

CLI-Hilfe:

```bash
grom --help
grom --version
```

## Nächste Schritte

- TLS-Profil und Speicher-Treiber wählen (`bbolt` für normale Installationen; `file` vor allem für Tests / winzige Instanzen) — [Konfiguration](configuration.md)
- Optional Passwort-Reset per E-Mail aktivieren (`mailer` + `auth.reset.public_base_url`) — [Konfiguration — Mailer und Passwort-Reset](configuration.md#mailer-und-passwort-reset)
- Optional ALTCHA-Captcha bei Registrierung/Login/„vergessen“ aktivieren (`auth.captcha.enabled`) — [Konfiguration — Auth-Captcha](configuration.md#auth-captcha-altcha)
- Selbstsignierte Zertifikate für lokales HTTPS erzeugen — `grom gencerts` (siehe TLS-Abschnitt in der Konfiguration)
- Produkt-Tour für den Client — [Benutzerüberblick](../user/overview.md)
- API-Referenz im Browser — `/api/docs/` auf dem laufenden Server
