# Sicherung und Wiederherstellung

Eine Serversicherung umfasst Metadaten und Dateien wie Tracks, Fotos, Avatare und Schlüssel. Sichern Sie beides zum selben Zeitpunkt.

## Sicherung erstellen {#backup}

1. Stoppen Sie Grom, damit Trainings und Importe während des Kopierens keine Daten ändern. Notieren Sie die Grom-Version und die vom Dienst verwendeten Pfade.
2. Kopieren Sie die Konfigurationsdatei und das gesamte Verzeichnis, auf das `storage.location` verweist. Es enthält die Dateien beider Treiber, die Föderationsschlüssel und das VAPID-Schlüsselpaar für Push-Benachrichtigungen.
3. Kopieren Sie bei `bbolt` auch die Datenbank, wenn `storage.bbolt.path` außerhalb von `storage.location` liegt. Standardmäßig liegt `grom.db` in diesem Verzeichnis. Sichern Sie bei statischem TLS auch das Zertifikat und den privaten Schlüssel aus der Konfiguration.
4. Bewahren Sie die Sicherung geschützt auf. Sie enthält private Trainingsdaten und Geheimnisse. Starten Sie Grom nach dem Kopieren wieder.

Liegt `storage.temp_dir` getrennt vom Datenverzeichnis, können temporäre Strava-Uploads außerhalb der Sicherung bleiben. Nach einer Unterbrechung können Benutzer dasselbe Archiv erneut hochladen.

## Wiederherstellen {#restore}

1. Stoppen Sie Grom und sichern Sie die aktuellen Daten separat, bevor Sie sie ersetzen.
2. Stellen Sie Konfiguration, Datenverzeichnis und externe Datenbank- oder TLS-Dateien aus derselben Sicherung wieder her. Verwenden Sie für die erste Wiederherstellung die ursprüngliche Grom-Version. Behalten Sie den Speichertreiber, den konfigurierten Föderationshostnamen und die Dateiberechtigungen bei. Prüfen Sie bei einem Wechsel des Binärverzeichnisses, wohin relative Pfade jetzt verweisen.
3. Starten Sie Grom mit der wiederhergestellten Konfiguration. Melden Sie sich an und öffnen Sie Ihre Trainingsliste sowie ein Training mit Track und Fotos. Prüfen Sie Track und Anhänge. Testen Sie die Wiederherstellung möglichst vorab in einer separaten Umgebung.

Treibereinstellungen und Migration sind unter [Konfiguration](configuration.md#storage-drivers) beschrieben.
