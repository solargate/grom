# Конфигурация

Grom настраивается YAML-файлом. По умолчанию ищет `config.yaml` в текущем рабочем каталоге. Передайте `--config` (или `-c`), чтобы указать другой путь.

- Примеры профилей: [`cmd/grom/config-examples/`](https://github.com/solargate/grom/tree/master/cmd/grom/config-examples)
- Все поля с комментариями: [`config.full.yaml`](https://github.com/solargate/grom/blob/master/cmd/grom/config-examples/config.full.yaml)

**Обязательно:** `auth.jwt_secret` — длинный случайный секрет для подписи JWT access-токенов.

## Основные параметры {#common-settings}

| Область | Что задать |
|---------|------------|
| `server.port` / `server.tls` | Порты прослушивания и режим TLS (`off`, `static` или `autocert`) |
| `server.registration` | Режим регистрации: `open` (по умолчанию), `closed` или `invite` |
| `storage.driver` / `location` / `temp_dir` | `file` (по умолчанию; тесты / крошечные инстансы) или `bbolt` (рекомендуется для обычных установок); корень данных и temp |
| `storage.bbolt.path` | Необязательный путь к `grom.db` при bbolt (по умолчанию: `{location}/grom.db`) |
| `federation.enabled` / `federation.domain` | ActivityPub; нужен HTTPS |
| `auth.reset` / `mailer` | Email для сброса пароля (`public_base_url`, SMTP или log-драйвер) |
| `auth.captcha` | Опциональный ALTCHA PoW на регистрацию/вход/«забыли» (`enabled`, опционально `hmac_secret` / `cost` / `expires_seconds`) |
| `logging.level` / `logging.format` | `debug`/`info`/`warn`/`error`; `text` (dev) или `json` (prod). По умолчанию: `info` + `json`. Отладочный вывод Gin (`[GIN-debug]`) включается только при `logging.level: debug`; иначе Gin в release mode |

Относительные пути в `storage.*`, `server.tls.cert_file` / `key_file`, `server.tls.autocert.cache_dir` и `federation.ca_cert_file` резолвятся относительно каталога бинарника `grom` (абсолютные пути используются как есть).

## Профили TLS {#tls-profiles}

| Профиль | Файл конфига | `tls.mode` | Федерация |
|---------|--------------|------------|-----------|
| Dev, только HTTP | `config.dev.notls.yaml` | `off` | выключена |
| Dev, self-signed TLS | `config.dev.tls.yaml` | `static` | включена |
| Prod, только HTTP | `config.prod.notls.yaml` | `off` | выключена |
| Prod, Let's Encrypt | `config.prod.tls.yaml` | `autocert` | включена |

**Dev со static TLS** — сгенерируйте сертификаты, затем запустите:

```bash
cd cmd/grom
go run . gencerts --ip 192.168.1.251 --domain 192.168.1.251
go run . --config config-examples/config.dev.tls.yaml
```

Для федерации между локальными инстансами задайте `federation.tls_insecure_skip_verify: true` и при необходимости `federation.ca_cert_file`, чтобы доверять вашему dev CA.

**Production с autocert** — нужен публичный DNS-имя в `federation.domain` (только hostname), порты **80** и **443** доступны из интернета. ACME-сертификаты по умолчанию кэшируются в `acme-cache` рядом с бинарником grom (переопределение: `server.tls.autocert.cache_dir`; используйте абсолютный путь, если бинарник лежит в системном каталоге вроде `/usr/bin`):

```bash
go run . --config config-examples/config.prod.tls.yaml
```

Замечания:

- Федерация требует HTTPS (`tls.mode: static` или `autocert`). С `tls.mode: off` она не работает.
- При включённой федерации лайки удалённых тренировок доставляются как ActivityPub `Like` / `Undo`, комментарии — как `Create`/`Delete` Note (`inReplyTo`); локальные тренировки принимают входящие лайки и комментарии с других инстансов. Лайки и комментарии на том же инстансе работают без федерации. Удаление аккаунта (`DELETE /api/v1/auth/me`) доставляет `Delete` локального актора в известные удалённые inbox (best-effort) перед очисткой локальных данных; inbox применяет удалённый actor `Delete`, очищая федеративный кэш этого владельца для получателя.
- Устаревшие конфиги с `server.tls.enabled: true` (без `mode`) трактуются как `mode: static`.

## Регистрация {#registration}

Параметр `server.registration` управляет возможностью регистрации новых пользователей:

| Значение | Поведение |
|----------|-----------|
| `open` | Регистрация открыта для всех (по умолчанию) |
| `closed` | Регистрация отключена; API возвращает 403 |
| `invite` | Регистрация только по приглашениям (механизм приглашений ещё не реализован; API возвращает 403) |

Текущий режим доступен через `GET /api/v1/server-info` в поле `registration`, чтобы клиент мог адаптировать интерфейс до попытки регистрации.

## Драйверы хранилища {#storage-drivers}

`file` подходит для тестов и очень маленьких инстансов (YAML-метаданные на диске удобно смотреть). Для обычной или production-установки предпочтителен `bbolt` — метаданные в Bolt DB, треки, фото и прочие blob остаются на файловой системе.

| Драйвер | Метаданные | Графики (скорость / пульс) | Лайки / комментарии тренировок | Blob (треки, фото, аватары, ключи) |
|---------|------------|----------------------------|-------------------------------|-------------------------------------|
| `file` (по умолчанию) | YAML под `storage.location` (`users.yaml`, per-user `equipment.yaml`, `profile.yaml`, YAML тренировок, …) | `speed-chart.json` и `heartrate-chart.json` в каталоге каждой тренировки | `likes.yaml` и `comments.yaml` рядом с каждой локальной тренировкой; федеративный кэш лайков/комментариев и ID outbox-activity в дереве `federation/` зрителя | То же дерево |
| `bbolt` | JSON в `{location}/grom.db` (или `storage.bbolt.path`); включая бакет `user_profiles` для UI-предпочтений | Упакованные бинарные значения в бакетах bbolt `speed_charts` / `fed_speed_charts` и `heart_rate_charts` / `fed_heart_rate_charts` (федеративный inbox) | Бакеты `workout_likes`, `fed_workout_likes`, `like_activities`, `workout_comments`, `fed_workout_comments` и `comment_activities` | Та же раскладка ФС под `storage.location` |

`postgres` зарезервирован в конфиге, но не реализован.

Миграция метаданных между драйверами (сначала остановите сервер; blob треков/медиа/аватаров общие и не копируются). Графики скорости и пульса конвертируются между JSON-blob файла и бинарными бакетами bbolt. Лайки и комментарии тренировок (локальные, федеративный кэш и исходящие activity id) и персональные токены копируются, чтобы оставаться читаемыми после смены драйвера:

| Копирует `migrate-storage` | Не копирует |
|----------------------------|-------------|
| Пользователи, профили, снаряжение, тренировки | Токены сброса пароля (короткоживущие; незавершённые ссылки сброса становятся недействительными) |
| Подписки, федеративные подписчики, авторы/тренировки федеративного inbox | Blob-файлы (треки, фото, аватары, ключи) — общие на диске |
| Локальные/федеративные лайки и комментарии + исходящие activity id | Временные задания Strava / captcha в памяти |
| Графики скорости/ЧСС (конвертация формата) | |
| Персональные токены доступа (PAT) | |

```bash
grom migrate-storage --config config.yaml --from file --to bbolt --verify
# затем задайте storage.driver: bbolt и перезапустите
grom migrate-storage --config config.yaml --from bbolt --to file --verify
```

`--dry-run` считает записи без записи; `--force` перезаписывает существующую базу bbolt.

Токены сброса пароля (`reset_tokens.yaml` / bbolt `reset_tokens`) короткоживущие и **не** копируются migrate-storage; незавершённые ссылки сброса после миграции недействительны. Устаревшие plain-text Like activity id (без `object_id`) восстанавливаются через федеративный inbox и, если задан `federation.domain`, через URL объектов локальных тренировок.

## Mailer и сброс пароля {#mailer-and-password-reset}

Исходящая почта опциональна. При `mailer.driver: off` (по умолчанию) сброс пароля выключен, и `GET /api/v1/server-info` сообщает `password_reset_enabled: false`.

| Параметр | Назначение |
|----------|------------|
| `auth.reset.public_base_url` | Базовый URL в ссылках сброса (без завершающего слэша). Обязателен, когда mailer включён. |
| `auth.reset.token_ttl_minutes` | Время жизни токена (по умолчанию `60`) |
| `mailer.driver` | `off`, `log` (писать в лог сервера — удобно в dev) или `smtp` |
| `mailer.from` | Адрес отправителя |
| `mailer.smtp.host` / `port` | SMTP-релей (обязателен при `driver: smtp`). Частые порты: `587` (STARTTLS), `465` (implicit TLS) |
| `mailer.smtp.username` / `password` | Необязательные учётные данные SMTP |
| `mailer.smtp.encryption` | `starttls` (по умолчанию; также по умолчанию для порта 587), `tls` (по умолчанию при порте `465`) или `none` |

**Нет** зависимости от локального MTA / `sendmail`: процесс говорит SMTP (через [go-mail](https://github.com/wneessen/go-mail)) с внешним провайдером (пароль приложения Gmail, SES, Mailgun и т.д.) или пишет сообщение в лог при `driver: log`.

Эндпоинты сброса пароля используют in-memory rate limiter с фиксированным окном (15 минут): forgot — 10 запросов на IP клиента и 3 на email; confirm reset — 20 на IP. Лимиты через Gin `ClientIP()` (учитывает `X-Forwarded-For` / `X-Real-IP`, если есть). Grom пока не раскрывает настройку trusted proxies, поэтому считайте forwarded-заголовки недоверенными, пока reverse proxy их не перезаписывает или не снимает.

Пример (production SMTP на порту 587):

```yaml
auth:
  jwt_secret: "..."
  reset:
    public_base_url: "https://grom.example.com"
mailer:
  driver: smtp
  from: "Grom <noreply@grom.example.com>"
  smtp:
    host: smtp.example.com
    port: 587
    username: "apikey"
    password: "secret"
    encryption: starttls
```

## Auth captcha (ALTCHA) {#auth-captcha-altcha}

Опциональная self-hosted [ALTCHA](https://altcha.org/) proof-of-work captcha (без стороннего сервиса и API-ключей). При `auth.captcha.enabled: true`:

- **Защищены:** `POST /api/v1/auth/register`, `/auth/login` и `/auth/password/forgot` — в теле запроса должен быть решённый payload `altcha` (base64 JSON).
- **Не защищены:** `POST /api/v1/auth/password/reset` (достаточно токена из письма).
- **Challenge:** `GET /api/v1/captcha/challenge` возвращает PoW-challenge (`200`). При выключенной captcha → `404`; при лимите challenge для IP → `429` с `Retry-After`.
- **Обнаружение клиентом:** `GET /api/v1/server-info` включает `captcha_enabled`. Flutter web/Android UI показывает чекбокс **I'm not a robot** / «Я не робот» и решает challenge локально перед отправкой.

| Параметр | Назначение |
|----------|------------|
| `auth.captcha.enabled` | Требовать captcha на регистрацию/вход/«забыли». По умолчанию: `false`. |
| `auth.captcha.hmac_secret` | HMAC-ключ для подписей challenge. Необязателен; если пусто, используется `auth.jwt_secret`. В production предпочтителен отдельный секрет, если JWT ротируете независимо. |
| `auth.captcha.cost` | Стоимость итераций PBKDF2 для PoW. По умолчанию: `1000`. Больше — тяжелее для CPU клиента. |
| `auth.captcha.expires_seconds` | Время жизни challenge в секундах. По умолчанию: `300`. |

Выдача challenge ограничена по rate in-memory (60 запросов на IP клиента за 15-минутное окно). Решённые payload одноразовые до истечения (replay отклоняется). Rate limits и replay-хранилище локальны для процесса — сбрасываются при перезапуске и не разделяются между несколькими процессами Grom; для multi-instance завершайте TLS/rate-limit на одном reverse proxy или держите captcha выключенной, пока нет sticky sessions / общего хранилища.

Разрешение IP клиента как у сброса пароля: Gin `ClientIP()` (учитывает `X-Forwarded-For` / `X-Real-IP`). Без доверенного reverse proxy, который перезаписывает эти заголовки, считайте их недоверенными.

Ошибки проверки возвращают `400` с сообщениями вроде `captcha is required`, `invalid captcha`, `captcha expired` или `captcha already used`.

Пример:

```yaml
auth:
  jwt_secret: "..."
  captcha:
    enabled: true
    # hmac_secret: "optional-separate-secret"
    cost: 1000
    expires_seconds: 300
```

## См. также {#see-also}

- [Установка и запуск](install.md)
- [Массовый импорт Strava](../integrations/strava-bulk-import.md) (`storage.temp_dir`)
