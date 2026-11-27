# Strava bulk import

This document describes how Grom imports workouts from a [Strava bulk data export](https://support.strava.com/hc/en-us/articles/216918437-Exporting-your-Data-and-Bulk-Export) ZIP archive.

## Overview

1. Upload a Strava export ZIP under **Integration → Strava**.
2. The server saves the archive in `{storage.temp_dir}/{nickname}/strava-{job_id}/` and reads it in place.
3. Grom prepares the available track and photos, resolves equipment, and saves each workout through the configured storage driver.
4. When the job finishes, Grom cleans up that job's directory.

If the server restarts mid-import, the job is reported as interrupted. Upload the same ZIP again. Grom skips activities already saved for your account.

## Configuration

```yaml
storage:
  temp_dir: tmp
```

Default: `tmp` (resolved relative to the grom binary, same as `storage.location`).

## activities.csv column mapping

Column numbers are **1-based**. Strava localizes CSV headers; Grom always uses indices.

| Col | Strava (RU example) | Grom field | Notes |
|-----|---------------------|------------|-------|
| 1 | ID физической активности | `external_id.id` | With `external_id.name` = `strava`; used for duplicate detection |
| 2 | Дата тренировки | `start_date` | Locale-aware date parsing |
| 3 | Название тренировки | `name` | |
| 4 | Тип активности | `sport_type` | Mapped to Grom sport type IDs |
| 5 | Описание | `description` | |
| 6 | Общее время | `duration_total_seconds` | Seconds |
| 9 | Относительное усилие | `relative_effort` | |
| 10 | Регулярный маршрут | `regular_track` | Boolean |
| 12 | Снаряжение | `equipment[]` | Resolved or created |
| 13 | Название файла | `track` | GPX or `.fit.gz` in `activities/` |
| 17 | Время в движении | `duration_seconds` | Seconds |
| 18 | Дистанция | `distance` | Meters |
| 19 | Макс. скорость | `speed_max_kmh` | CSV is m/s; Grom multiplies by 3.6 → km/h |
| 20 | Средняя скорость | `speed_avg_kmh` | CSV is m/s; Grom multiplies by 3.6 → km/h |
| 21 | Набор высоты | `elevation_gain` | Meters |
| 22 | Высота спуска | `elevation_loss` | Meters |
| 23 | Высота низменности | `elevation_low` | Meters |
| 24 | Высота подъема | `elevation_high` | Meters |
| 25 | Макс. уклон | `grade_max` | |
| 26 | Средний угол уклона | `grade_avg` | |
| 29 | Макс. каденс | `cadence_max` | |
| 30 | Средний каденс | `cadence_avg` | |
| 31 | Макс. пульс | `heart_rate_max` | |
| 32 | Средний пульс | `heart_rate_avg` | |
| 33 | Макс. мощность | `watts_max` | Watts |
| 34 | Средняя мощность | `watts_avg` | Watts |
| 35 | Калории | `calories` | |
| 36 | Макс. температура | `temperature_max` | °C |
| 37 | Средняя температура | `temperature_avg` | °C |
| 86 | Всего шагов | `steps_total` | |
| 93 | Всего циклов | `cycles_total` | |
| 101 | Всего подходов | `sets_total` | |
| 102 | Общее количество повторений | `reps_total` | |
| 103 | Медиафайл | `media_files` | Photos from `media/`, videos ignored |

Ignored columns include col 7 (localized distance in km) — use col 18 (meters) instead.

<a id="track-handling-attachtrack-vs-createwithtrack"></a>

## Track handling {#track-handling}

| Method | Use case | Overwrites client metrics |
|--------|----------|---------------------------|
| `CreateWithTrack` | Manual / multipart create with a track (including Strava API sync) | **No** — non-empty client `start_date`, durations, `distance`, speeds, and elevation fields are preserved; the track fills only empty metrics |
| `ImportWithAttachments` | Strava ZIP import | No — CSV values are preserved |

Both paths can set the device name from FIT and generate a map preview when GPS data is available.

## Equipment

`activities.csv` (col 12) stores the gear display name, while `bikes.csv` and `shoes.csv`
keep the parts in separate columns (col 1 = nickname, col 2 = brand, col 3 = model). The
nickname column is empty when the gear has no custom name, so Grom matches on every
combination of these parts instead of the nickname alone.

1. Match activity equipment name (col 12) against the user's existing equipment.
2. If not found, look up `bikes.csv` → create `bike`.
3. Else look up `shoes.csv` → create `shoes`.
4. Else create `other` equipment with the given name.

Gear from `bikes.csv` / `shoes.csv` is created with the nickname as name (falling back to
`brand model`), plus brand and model.

## Locale detection

Grom detects export locale from sport type names and date strings, then parses:

- **Dates** — English and Russian Strava formats, including four-letter Russian month abbreviations (`сент.`, `нояб.`, `февр.`).
- **Numbers** — comma or dot decimal separators, thousand separators.
- **Booleans** — `true` / `false` (locale-independent in Strava exports).

## Sport type mapping

Localized Strava sport names are mapped to Grom sport type IDs (`Run`, `Ride`, `WeightTraining`, …). When Strava exports a generic type such as `Workout` / `Тренировка`, Grom may infer a more specific sport from the activity name (for example `Пилатес (день)` → `Pilates`). Unknown types fall back to `Workout`.

Import results include:

- `parse_skipped` — rows from `activities.csv` that could not be parsed (for example invalid dates) and were not imported
- `media_missing` — photo paths listed in the CSV media column that were not present in the ZIP (Strava bulk export often omits older media; videos are ignored and not counted)

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/integrations/strava/import` | Upload ZIP (`archive` field), returns `202` |
| `GET` | `/api/v1/integrations/strava/import/status` | Upload/import progress |

## Duplicate import

If a workout with the same `external_id` (`name` = `strava` and matching `id`) already exists for the user, the activity is skipped.

## Related

- [Strava API import (Android)](strava-api-import.md)
- [Import tracks](import-tracks.md)
- [User overview](../user/overview.md)
