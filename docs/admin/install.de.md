# Installation und Start

Installieren Sie Grom für einen schnellen Einstieg aus einem GitHub-Release. Die Weboberfläche ist in der Server-Binärdatei enthalten. Einstellungen für öffentlich erreichbare Server finden Sie unter [Konfiguration](configuration.md).

<a id="download-von-github-releases"></a>

## Download von GitHub Releases {#download-from-github-releases}

Vorgefertigte Pakete erscheinen mit jedem [GitHub-Release](https://github.com/solargate/grom/releases):

| Asset | Inhalt |
|-------|--------|
| `grom-<version>-linux-amd64.tar.gz` | Server-Binary + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Server-Binary + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Server-Binary + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Server-Binary + `config-examples/` |

Dasselbe Release enthält auch den Android-Client als `grom-<version>.apk`.

<a id="start"></a>

## Starten {#run}

1. Laden Sie das Serverarchiv für Ihr Betriebssystem und Ihre Prozessorarchitektur herunter und entpacken Sie es. Öffnen Sie ein Terminal im entpackten Verzeichnis mit der Binärdatei und `config-examples/`.
2. Kopieren Sie die Vorlage für lokales HTTP nach `config.yaml` und setzen Sie einen langen, zufälligen Wert für `auth.jwt_secret`.
3. Starten Sie den Server mit dieser Datei.

**Linux und macOS**

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# Vor dem Start auth.jwt_secret in config.yaml setzen.
./grom --config config.yaml
```

**Windows PowerShell**

```powershell
Copy-Item config-examples/config.dev.notls.yaml config.yaml
# Vor dem Start auth.jwt_secret in config.yaml setzen.
.\grom.exe --config config.yaml
```

Öffnen Sie `http://localhost:8080/` im Browser. Registrieren Sie ein Konto und melden Sie sich an. Das Beispiel verwendet HTTP für einen lokalen Test. Richten Sie [TLS](configuration.md#tls-profiles) ein, bevor Sie den Server öffentlich erreichbar machen.

Ohne `--config` (oder `-c`) sucht Grom nach `config.yaml` im aktuellen Arbeitsverzeichnis. Relative Speicherpfade beziehen sich auf das Verzeichnis der Binärdatei.

Der Android-Client kann sich mit derselben Instanz verbinden, wenn der Server vom Gerät aus erreichbar ist. Geben Sie die aus Sicht des Geräts erreichbare Adresse ein. Die Wahl der Serveradresse und die Anmeldung beschreibt der [Benutzerüberblick](../user/overview.md). Die Swagger-Oberfläche finden Sie auf dem laufenden Server unter `/api/docs/`.

## Nächste Schritte

- TLS-Profil und Speicher-Treiber wählen (`bbolt` für normale Installationen; `file` vor allem für Tests / winzige Instanzen) — [Konfiguration](configuration.md)
- Sichern Sie die Serverdaten und testen Sie die Wiederherstellung — [Sicherung und Wiederherstellung](backup-restore.md)
- Optional Passwort-Reset per E-Mail aktivieren (`mailer` + `auth.reset.public_base_url`) — [Konfiguration — Mailer und Passwort-Reset](configuration.md#mailer-und-passwort-reset)
- Optional ALTCHA-Captcha bei Registrierung/Login/„vergessen“ aktivieren (`auth.captcha.enabled`) — [Konfiguration — Auth-Captcha](configuration.md#auth-captcha-altcha)
- Selbstsignierte Zertifikate für lokales HTTPS erzeugen — `grom gencerts` (siehe TLS-Abschnitt in der Konfiguration)
- Produkt-Tour für den Client — [Benutzerüberblick](../user/overview.md)
- API-Referenz im Browser — `/api/docs/` auf dem laufenden Server

<a id="aus-dem-quellcode-bauen"></a>

## Aus dem Quellcode bauen {#build-from-source}

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

Wechseln Sie nach dem Build nach `cmd/grom` und folgen Sie der [Startanleitung](#run) mit der Binärdatei und den Konfigurationsbeispielen aus diesem Verzeichnis.
