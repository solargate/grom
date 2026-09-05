# Benutzerüberblick

Der Flutter-Client von Grom läuft als **Web-UI** und als **Android**-App. Die Web-UI liefert derselbe `grom`-Prozess: öffnen Sie die Basis-URL des Servers im Browser (z. B. `http://localhost:8080/` mit der Standard-Dev-Konfiguration). Bildschirme und Abläufe entsprechen der Android-App; Live-GPS-Aufzeichnung gibt es nur unter Android. UI-Texte gibt es auf Englisch, Russisch und Deutsch.

Wenn Sie nicht angemeldet sind, zeigt **Home** / „Startseite“ einen Willkommensbildschirm mit Grom-Logo, kurzer Beschreibung und den Schaltflächen **Sign in** / „Anmelden“ und **Register** / „Registrieren“. Unter Android erinnert der Text zusätzlich daran, beim Anmelden die Adresse des Grom-Servers einzugeben.

Unter **Android** (später auch iOS) fragen Anmeldung und Registrierung nach einer **Server-URL**. Sie können einen nackten Host wie `grom.example.com` eingeben (ohne `https://`) oder neben dem Feld die Liste öffnen und einen [freigegebenen öffentlichen Server](approved-servers.md) oder einen bereits genutzten wählen. Beim Absenden prüft die App `GET /api/v1/status` zuerst über HTTPS, dann HTTP, schreibt die aufgelöste URL ins Feld und fährt fort. Wenn Sie bereits `http://` / `https://` oder einen expliziten Port eingeben, wird dieser Wert unverändert genutzt. **HTTP ist für lokale / LAN-Instanzen** ohne TLS erlaubt; für öffentlich erreichbare Server bevorzugen Sie HTTPS.

## Anmeldung und Passwort-Reset

Wenn der Betreiber ausgehende E-Mail aktiviert (`mailer` in der Server-Konfiguration), zeigt der Anmeldebildschirm **Forgot password?** / „Passwort vergessen?“. Geben Sie die Konto-E-Mail ein; der Server antwortet immer gleich, ob die Adresse registriert ist oder nicht. Prüfen Sie den Posteingang und öffnen Sie den Reset-Link im **Browser** (Web-UI unter `/reset-password`). Nach dem Setzen eines neuen Passworts melden Sie sich erneut in der App oder im Web an. Passwort-Reset ist nicht verfügbar, wenn der Server `password_reset_enabled: false` meldet.

Wenn der Betreiber Captcha aktiviert (`auth.captcha.enabled`), zeigen Anmeldung, Registrierung und „Passwort vergessen“ eine Checkbox **I'm not a robot** / „Ich bin kein Roboter“. Aktivieren Sie sie und warten Sie, bis die lokale Proof-of-Work-Prüfung fertig ist, bevor Sie absenden (auf langsameren Geräten kann das einen Moment dauern). Das Setzen eines neuen Passworts über den E-Mail-Reset-Link erfordert kein Captcha.

## Profil und Kontolöschung

Unter **Profile** / „Profil“ können Sie über das Overflow-Menü das Profil bearbeiten oder **Delete account** / „Konto löschen“ wählen. Vollständige Schritte, welche Daten entfernt werden und Föderationshinweise: [Grom-Konto löschen](delete-account.md).

Diese Seite ist eine kurze Tour der wichtigsten Bildschirme (Screenshots unten von Android). Admin-Setup (Installation, Config, TLS, Föderation) steht unter [Admin-Dokumentation](../README.md#administration). Für die HTTP-API siehe Swagger unter `/api/docs/` auf dem laufenden Server.

## Workouts

Ihr Home-Feed listet Aktivitäten mit Typ, Datum, Gerät, Distanz/Zeit (und Tempo oder Höhenmeter, wenn relevant) sowie einer Kartenvorschau, wenn ein GPS-Track angehängt ist. Auf dem Tab **My workouts** / „Meine Trainings“ können Sie die Liste nach Sportart filtern: das Filter-Icon in der App-Leiste (nur auf diesem Tab, wenn Sie bereits Sportarten genutzt haben) öffnet Umschalter für diese Sportarten. Derselbe Tab hat einen Umschalter Liste/Karten (neben dem Filter, wenn sichtbar): kompakte Zeilen zeigen Sport-Icon, Datum, Name sowie Distanz oder Dauer je nach Sportkategorie; die Wahl bleibt auf dem Gerät gespeichert. Jede Karte zeigt unter Karte und Fotos eine Social-Leiste: links Likes (Daumen hoch zum Liken/Entliken fremder Workouts; Zähler tippen für Likes-Liste) und rechts Kommentare (Zähler + Kommentar-Icon). Eigene Workouts können Sie nicht liken; der Like-Button bleibt bei Ihren deaktiviert. Kommentieren dürfen alle eigene und fremde Workouts. Tippen Sie auf die Kommentar-Steuerung, um den Thread zu öffnen, einen Kommentar hinzuzufügen (bis 1000 Zeichen) oder einen selbst geschriebenen Kommentar zu löschen (Workout-Besitzer können auch beliebige Kommentare an ihrem Workout löschen).

![Workout-Liste auf Android](../screenshots/workout-list.jpg)

Öffnen Sie ein Workout für die volle Karte: interaktive Karte (wenn ein Track vorhanden ist), Fotogalerie, dieselbe Social-Leiste wie in der Liste, ein Geschwindigkeits-über-Distanz-Diagramm mit Durchschnitt und Maximum sowie ein Herzfrequenz-Diagramm (Distanz bei GPS, sonst vergangene Minuten) mit Durchschnitt und Maximum. Tippen Sie auf ein Diagramm, um Werte an diesem Punkt zu sehen.

Sie können Workouts manuell anlegen, GPX/FIT-Tracks importieren, Fotos anhängen und Ausrüstung verknüpfen. Beim Bearbeiten können Sie Fotos hinzufügen oder entfernen (bis 20). Beim Anlegen ist der Workout-Name standardmäßig der lokalisierte Sporttyp und folgt Sportwechseln, bis Sie ihn bearbeiten. Unter Android können Sie auch einen Live-GPS-Track aufzeichnen (siehe unten).

## Live-Aufzeichnung (Android)

Öffnen Sie **Add workout → Record** / „Workout hinzufügen → Aufzeichnen“, um eine Route auf der Karte mit Live-Dauer, Geschwindigkeit und Distanz zu tracken. Lassen Sie die Aufzeichnungs-Benachrichtigung aktiv, solange die Sitzung läuft. Auto-Pause lässt sich ein- oder ausschalten.

![Workout-Aufzeichnung auf Android](../screenshots/workout-record.jpg)

## Ausrüstung

Verwalten Sie Fahrräder, Schuhe und anderes Gear. Einträge sind nach Kategorie gruppiert; Distanzsummen wachsen, wenn Sie Gear in Workouts nutzen (auch nach Strava-Massenimport).

![Ausrüstungsliste auf Android](../screenshots/equipment.jpg)

## Social und Föderation

Folgen Sie anderen Nutzern derselben Instanz und browsen Sie einen gemeinsamen Feed. Liken und kommentieren Sie Workouts von Personen, denen Sie folgen (lokal oder föderiert); Zähler erscheinen in Listen- und Detailansichten.

Wenn der Betreiber ActivityPub-Föderation aktiviert, können Sie auch Sportlern auf anderen Grom-Instanzen folgen (HTTPS auf dem Server erforderlich). Likes auf Remote-Workouts werden als ActivityPub-`Like` gesendet; Entfernen eines Likes sendet `Undo`. Kommentare zu Remote-Workouts gehen als `Create` Note mit `inReplyTo`; Löschen eines Kommentars sendet `Delete`. Eingehende Likes und Kommentare von anderen Instanzen aktualisieren das lokale Workout genauso. Kontolöschung und föderiertes Actor-`Delete` sind in [Grom-Konto löschen](delete-account.md) beschrieben.

## Strava-Import

Unter **Integration → Strava** laden Sie ein ZIP des [Strava-Massenexports](https://support.strava.com/hc/en-us/articles/216918437-Exporting-your-Data-and-Bulk-Export) hoch. Grom importiert Aktivitäten, Tracks, Fotos und Ausrüstung. Spaltenzuordnung und serverseitiges Verhalten: [Strava-Massenimport](../integrations/strava-bulk-import.md).

Nur unter **Android** können Sie zusätzlich **Trainings aus Strava importieren** mit eigener Strava-API-Client-ID/Secret aktivieren, einmal per OAuth verbinden und den Sync-Button auf Start nutzen (bis zu 10 neueste Aktivitäten, Stopp bei der ersten bereits importierten). Details: [Strava-API-Import](../integrations/strava-api-import.md).

## Tracks importieren

Unter **Integration** → **External services** öffnet **Tracks importieren** den System-Dateidialog für eine oder mehrere `.gpx`-/`.fit`-Dateien (Web und Android). Details und `external_id`-Dedup: [Tracks importieren](../integrations/import-tracks.md).
