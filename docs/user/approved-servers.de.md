# Freigegebene Grom-Server

Die Android-App kann sich mit **jeder** Grom-Instanz verbinden: Geben Sie die Server-URL bei Anmeldung, Registrierung oder Passwort-Reset ein. Sie können auch aus einer Liste **freigegebener öffentlicher Server** wählen, die in der App mitgeliefert wird, plus Server, die Sie bereits genutzt haben.

Diese Seite richtet sich an:

- Nutzer der Android-App, die eine öffentliche Instanz wählen möchten
- Betreiber, die ihre öffentliche Instanz in diesem Auswahlfeld listen lassen wollen

Ein Listeneintrag ist **keine** Empfehlung der Datenschutzpraxis oder Verfügbarkeit des Betreibers. Grom hat keine zentrale Cloud; Sie wählen, auf welchem Server Ihr Konto liegt. Siehe [Privacy Policy](../privacy.md).

## In der Android-App

Auf **Anmelden**, **Registrieren** und **Passwort vergessen** hat das Server-URL-Feld eine Dropdown-Schaltfläche.

- **Freigegebene Server** stammen aus dem Katalog, der in diese App-Version einkompiliert ist: URL (fett), Name und eine kurze Beschreibung.
- **Zuletzt verwendete Server** sind URLs, die Sie selbst eingegeben und danach **erfolgreich angemeldet oder registriert** haben. Sie stehen unter der freigegebenen Liste. Passwort-Reset fügt keinen Server hinzu. Es gibt keine Löschfunktion; Einträge rutschen nach vielen weiteren erfolgreichen Verbindungen heraus.
- Sie können weiterhin einen beliebigen Host oder URL eingeben (Schema optional). HTTP ist für lokale / LAN-Instanzen erlaubt. Wie der Client HTTPS und HTTP prüft, steht im [Benutzerüberblick](overview.md).

Der Katalog ist **im APK enthalten**. Eine Änderung im Git aktualisiert bereits installierte Apps nicht; neue Einträge erscheinen nach Installation eines Android-Releases mit dieser Änderung.

## Instanz hinzufügen (Pull Request)

Jede Person kann einen Pull Request gegen [`server-catalog.yaml`](https://github.com/solargate/grom/blob/master/server-catalog.yaml) im Repository-Root öffnen.

1. Forken Sie [solargate/grom](https://github.com/solargate/grom) und legen Sie einen Branch an.
2. Fügen Sie **einen** Eintrag unter `servers:` hinzu (andere Instanzen nicht entfernen oder umschreiben).
3. Öffnen Sie einen Pull Request nach `master`. Der Maintainer prüft ihn.

Beispiel:

```yaml
servers:
  - url: https://grom.example.org
    name: Example Grom
    description: Public instance for testing federation.
    email: admin@example.org
```

### Felder

| Feld | In der App | Regeln |
|------|------------|--------|
| `url` | Ja (fett) | Muss mit `https://` beginnen. **Kein Port** (auch nicht `:443`). Ein Pfad ist erlaubt (`https://host/grom`). Kein Userinfo, Query oder Fragment. Hostname muss ein DNS-Name sein (keine IP). |
| `name` | Ja | Kurzer Anzeigename, höchstens 80 Zeichen. Eine Sprache, so wie Sie sie schreiben. |
| `description` | Ja (kleine Schrift) | Ein bis zwei Sätze, höchstens 280 Zeichen. Eine Sprache, so wie Sie sie schreiben. |
| `email` | Nein | Kontakt des Betreibers für das Review. Muss `@` enthalten. |

CI lehnt ungültiges YAML ab (kein `https://`, Ports, Duplikate nach Normalisierung von Schrägstrichen usw.). Maintainer können Einträge außerdem ablehnen, die keine öffentliche Grom-Instanz sind, ein defektes TLS-Zertifikat haben oder irreführend wirken.

Nach dem Merge erscheint die Instanz im **nächsten Android-Release**. Führen Sie `make catalog` aus, wenn in derselben Änderung die generierte Flutter-Datei aktualisiert werden soll; CI schlägt fehl, wenn die Datei veraltet ist.

## Siehe auch

- [Benutzerüberblick — Anmeldung](overview.md#anmeldung-und-passwort-reset)
- [Installation und Start](../admin/install.md) (Android mit dem eigenen Server verbinden)
