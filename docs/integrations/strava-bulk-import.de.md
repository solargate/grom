# Strava-Massenimport

Dieses Dokument beschreibt, wie Grom Workouts aus einem ZIP des [Strava-Massendatenexports](https://support.strava.com/hc/en-us/articles/216918437-Exporting-your-Data-and-Bulk-Export) importiert.

## Überblick

1. Der Nutzer lädt ein Strava-Export-ZIP auf der Seite **Integration → Strava** hoch.
2. Das Archiv wird unter `{storage.temp_dir}/{nickname}/` gespeichert.
3. Der Server öffnet das ZIP mit Gos `archive/zip` und liest Dateien direkt (keine volle Extraktion).
4. Pro Aktivität erzeugt Grom ein Workout-YAML und hängt optional Track (`AttachTrack`), Fotos und Ausrüstung an.
5. Nach Abschluss wird das temporäre Importverzeichnis des Nutzers entfernt.

## Konfiguration

```yaml
storage:
  temp_dir: tmp
```

Standard: `tmp` (relativ zur grom-Binary, wie `storage.location`).

## Spaltenzuordnung activities.csv

Spaltennummern sind **1-basiert**. Strava lokalisiert CSV-Header; Grom nutzt immer Indizes.

| Col | Strava (RU-Beispiel) | Grom-Feld | Hinweise |
|-----|----------------------|-----------|----------|
| 1 | ID физической активности | `external_id.id` | Mit `external_id.name` = `strava`; für Duplikaterkennung |
| 2 | Дата тренировки | `start_date` | Locale-bewusstes Datumsparsing |
| 3 | Название тренировки | `name` | |
| 4 | Тип активности | `sport_type` | Auf Grom-Sporttyp-IDs gemappt |
| 5 | Описание | `description` | |
| 6 | Общее время | `duration_total_seconds` | Sekunden |
| 9 | Относительное усилие | `relative_effort` | |
| 10 | Регулярный маршрут | `regular_track` | Boolean |
| 12 | Снаряжение | `equipment[]` | Aufgelöst oder erzeugt |
| 13 | Название файла | `track` | GPX oder `.fit.gz` in `activities/` |
| 17 | Время в движении | `duration_seconds` | Sekunden |
| 18 | Дистанция | `distance` | Meter |
| 19 | Макс. скорость | `speed_max_kmh` | CSV ist m/s; Grom multipliziert mit 3.6 → km/h |
| 20 | Средняя скорость | `speed_avg_kmh` | CSV ist m/s; Grom multipliziert mit 3.6 → km/h |
| 21 | Набор высоты | `elevation_gain` | Meter |
| 22 | Высота спуска | `elevation_loss` | Meter |
| 23 | Высота низменности | `elevation_low` | Meter |
| 24 | Высота подъема | `elevation_high` | Meter |
| 25 | Макс. уклон | `grade_max` | |
| 26 | Средний угол уклона | `grade_avg` | |
| 29 | Макс. каденс | `cadence_max` | |
| 30 | Средний каденс | `cadence_avg` | |
| 31 | Макс. пульс | `heart_rate_max` | |
| 32 | Средний пульс | `heart_rate_avg` | |
| 33 | Макс. мощность | `watts_max` | Watt |
| 34 | Средняя мощность | `watts_avg` | Watt |
| 35 | Калории | `calories` | |
| 36 | Макс. температура | `temperature_max` | °C |
| 37 | Средняя температура | `temperature_avg` | °C |
| 86 | Всего шагов | `steps_total` | |
| 93 | Всего циклов | `cycles_total` | |
| 101 | Всего подходов | `sets_total` | |
| 102 | Общее количество повторений | `reps_total` | |
| 103 | Медиафайл | `media_files` | Fotos aus `media/`, Videos werden ignoriert |

Alle neuen YAML-Felder werden **nach `distance`** und vor `track` geschrieben.

Ignorierte Spalten umfassen Col 7 (lokalisierte Distanz in km) — stattdessen Col 18 (Meter) nutzen.

## Track-Handling: `AttachTrack` vs `CreateWithTrack`

| Methode | Anwendungsfall | Überschreibt CSV-Metriken |
|---------|----------------|---------------------------|
| `CreateWithTrack` | Manueller Workout-Upload in der UI | Ja (`start_date`, `duration_seconds`, `distance`) |
| `AttachTrack` | Strava-Import | **Nein** — CSV-Werte bleiben erhalten |

`AttachTrack` schreibt trotzdem die Track-Datei, setzt `device` aus FIT wenn verfügbar und erzeugt `map-preview.webp`, wenn GPS-Daten vorhanden sind.

## Ausrüstung

`activities.csv` (Col 12) enthält den Anzeigenamen der Ausrüstung, während `bikes.csv`
und `shoes.csv` die Bestandteile in getrennten Spalten führen (Col 1 = Spitzname,
Col 2 = Marke, Col 3 = Modell). Die Spitzname-Spalte ist leer, wenn die Ausrüstung keinen
eigenen Namen hat, deshalb gleicht Grom alle Kombinationen dieser Bestandteile ab und
nicht nur den Spitznamen.

1. Aktivitäts-Ausrüstungsname (Col 12) mit vorhandener Ausrüstung des Nutzers abgleichen.
2. Falls nicht gefunden, in `bikes.csv` nachschlagen → `bike` erzeugen.
3. Sonst in `shoes.csv` nachschlagen → `shoes` erzeugen.
4. Sonst `other`-Ausrüstung mit dem gegebenen Namen erzeugen.

Ausrüstung aus `bikes.csv` / `shoes.csv` wird mit dem Spitznamen als Name angelegt (sonst
`Marke Modell`), zusätzlich mit Marke und Modell.

## Locale-Erkennung

Grom erkennt die Export-Locale anhand von Sporttyp-Namen und Datumsstrings und parst dann:

- **Daten** — englische und russische Strava-Formate, inklusive vierbuchstabiger russischer Monatsabkürzungen (`сент.`, `нояб.`, `февр.`).
- **Zahlen** — Komma- oder Punkt-Dezimaltrenner, Tausendertrenner.
- **Booleans** — `true` / `false` (in Strava-Exporten locale-unabhängig).

## Sporttyp-Mapping

Lokalisierte Strava-Sportnamen werden auf Grom-Sporttyp-IDs gemappt (`Run`, `Ride`, `WeightTraining`, …). Wenn Strava einen generischen Typ wie `Workout` / `Тренировка` exportiert, kann Grom aus dem Aktivitätsnamen einen spezifischeren Sport ableiten (z. B. `Пилатес (день)` → `Pilates`). Unbekannte Typen fallen auf `Workout` zurück.

Importergebnisse enthalten:

- `parse_skipped` — Zeilen aus `activities.csv`, die nicht geparst werden konnten (z. B. ungültige Daten) und nicht importiert wurden
- `media_missing` — Foto-Pfade aus der CSV-Media-Spalte, die im ZIP fehlten (Strava-Massenexport lässt ältere Medien oft weg; Videos werden ignoriert und nicht gezählt)

## API

| Method | Path | Beschreibung |
|--------|------|--------------|
| `POST` | `/api/v1/integrations/strava/import` | ZIP hochladen (Feld `archive`), Antwort `202` |
| `GET` | `/api/v1/integrations/strava/import/status` | Upload-/Importfortschritt |

## Doppelter Import

Existiert für den Nutzer bereits ein Workout mit derselben `external_id` (`name` = `strava` und passende `id`), wird die Aktivität übersprungen.

## Verwandt

- [Strava-API-Import (Android)](strava-api-import.md)
- [Tracks importieren](import-tracks.md)
- [Benutzerüberblick](../user/overview.md)
