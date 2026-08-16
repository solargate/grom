# Grom-Konto löschen

Diese Seite erklärt, wie Sie Ihr **Grom**-Konto und die dafür auf einem Grom-Server gespeicherten Daten dauerhaft löschen. Nutzen Sie sie, wenn Sie kein Konto mehr auf einer Instanz wollen, bei der Sie sich registriert haben (auch nach dem Deinstallieren der Android-App).

Grom ist self-hosted: Android-Client und Web-UI verbinden sich mit **der Server-URL, die Sie konfiguriert haben**. Konto und Workouts liegen auf dieser Instanz. Die Löschung läuft auf **diesem** Server — es gibt keine zentrale Grom-Cloud mit den Daten aller Nutzer.

## Löschen über die Web-UI (ohne Android-Neuinstallation)

1. Öffnen Sie die Basis-URL Ihrer Instanz im Browser (derselbe Host wie im Server-Feld der Android-App, z. B. `https://grom.example.com/`).
2. Melden Sie sich mit Nickname (oder E-Mail) und Passwort an.
3. Öffnen Sie **Profile** / „Profil“.
4. Öffnen Sie das Overflow-Menü (⋮) und wählen Sie **Delete account** / „Konto löschen“.
5. Lesen Sie die Warnung, geben Sie Ihr Passwort ein und bestätigen Sie **Delete** / „Löschen“.
6. Der Client zeigt **Goodbye** / „Auf Wiedersehen“ und meldet Sie ab. Das Konto ist auf diesem Server weg.

## Löschen in der Android-App

Die Schritte entsprechen der Web-UI: **Profile** → Overflow-Menü → **Delete account** → Passwort eingeben → bestätigen. Sie müssen bei der Instanz angemeldet sein, auf der das Konto existiert.

## Was gelöscht wird

Nach erfolgreicher Bestätigung entfernt der Server dauerhaft die mit Ihrem Konto verbundenen Daten auf dieser Instanz, einschließlich:

- Login-Credentials, Profil und Avatar
- Ihre Workouts (Metadaten, GPS-Tracks, Medien, Kartenvorschauen, Geschwindigkeits- und Herzfrequenz-Diagramme)
- Ausrüstung
- Follow-Beziehungen, an denen Sie beteiligt sind
- Ihre Likes und Kommentare an Workouts anderer Nutzer auf dieser Instanz
- Persönliche Zugriffstoken (PAT)
- Passwort-Reset-Token
- Lokale Föderationsdaten für Ihren Actor (Schlüssel, Follower-/Outbox-Zustand für Ihren Nickname)

Derselbe Nickname und dieselbe E-Mail können auf dieser Instanz sofort erneut registriert werden.

## Wann die Löschung erfolgt

Die Löschung ist **sofort** nach erfolgreicher Passwortbestätigung. Es gibt in Grom keine verzögerte Purge-Warteschlange.

## Föderation

Wenn die Instanz ActivityPub-Föderation aktiviert hat, sendet Grom vor dem Löschen lokaler Daten ein `Delete` Ihres Actors an bekannte Remote-Inboxes (**best-effort**). Remote-Server, die die Activity nie erhalten, können veraltete Kopien Ihres föderierten Inhalts behalten. Empfängt dieser Server ein Remote-Actor-`Delete`, werden die föderierten Inbox-Einträge dieses Sportlers für lokale Nutzer gelöscht.

## Anforderungen und Grenzen

- Sie müssen mit einer normalen Sitzung (JWT) angemeldet sein. Persönliche Zugriffstoken **können** kein Konto löschen.
- Ihr aktuelles Passwort ist erforderlich.
- Auf dieser Dokumentationsseite gibt es keinen alternativen Anfragekanal: öffnen Sie die Web-UI **Ihrer** Instanz und nutzen Sie **Delete account** wie oben.
