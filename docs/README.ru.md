# Документация

Grom — self-hosted трекер тренировок с опциональным слоем федерации ActivityPub. Записывайте или импортируйте тренировки, управляйте снаряжением, подписывайтесь на других спортсменов (локальных или федеративных), ставьте лайки и комментируйте активности и просматривайте социальную ленту — на инфраструктуре, которую контролируете вы.

Сервер — один бинарник на Go; Flutter-клиент поставляется как встроенный веб-UI и как Android-приложение.

## Возможности {#features}

- **Тренировки** — создание и редактирование активностей со статистикой, заметками, медиа и превью карты
- **GPS-треки** — импорт GPX и FIT; запись в реальном времени на Android
- **Снаряжение** — велосипеды, обувь и другое снаряжение, привязанное к тренировкам
- **Социальная лента** — подписки на пользователей и их тренировки в одной ленте
- **Лайки** — лайк чужих активностей, счётчики и список лайкнувших; федеративные `Like` / `Undo` при включённом ActivityPub
- **Комментарии** — комментарии к своим и чужим активностям (добавление/список/удаление); федеративные `Create`/`Delete` Note при включённом ActivityPub
- **Федерация** — опциональный ActivityPub, чтобы инстансы могли подписываться друг на друга
- **Импорт Strava** — массовый импорт ZIP-экспорта данных Strava
- **Health Sync** — импорт активностей из Google Drive на Android (экспорт Health Sync)
- **API-токены** — персональные токены с областями доступа (`grom_pat_…`) для тренировок и снаряжения
- **Клиенты** — один Flutter UI в браузере (отдаётся сервером) и как Android APK
- **Локали** — английский, русский и немецкий в Flutter UI

<p align="center">
  <img src="screenshots/workout-list.jpg" width="250" alt="Список тренировок" />
  <img src="screenshots/workout-record.jpg" width="250" alt="Запись тренировки" />
  <img src="screenshots/equipment.jpg" width="250" alt="Снаряжение" />
</p>

Обзор экранов: [Обзор для пользователя](user/overview.md). Сборка и запуск: [Установка и запуск](admin/install.md). Исходники и быстрый старт: [README репозитория](https://github.com/solargate/grom#readme).

## Я хочу… {#i-want-to}

### Пользователь {#user}

| Цель | Страница |
|------|----------|
| Узнать, что умеет клиент (тренировки, лайки, комментарии, запись, снаряжение) | [Обзор для пользователя](user/overview.md) |
| Сбросить забытый пароль (если оператор включил почту) | [Обзор — Вход и сброс пароля](user/overview.md#sign-in-and-password-reset) |
| Войти / зарегистрироваться, когда на инстансе включена captcha | [Обзор — Вход и сброс пароля](user/overview.md#sign-in-and-password-reset) |
| Пользоваться Grom в браузере (тот же UI, что на Android) | Откройте базовый URL сервера после [установки](admin/install.md); см. [Обзор для пользователя](user/overview.md) |
| Импортировать экспорт Strava (UI + как работает импорт) | [Массовый импорт Strava](integrations/strava-bulk-import.md) |
| Импортировать активности Health Sync из Google Drive (Android) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Создать API-токены для скриптов и внешних приложений | [API-токены Grom](user/grom-api-tokens.md) |
| Удалить свой аккаунт | [Удаление аккаунта Grom](user/delete-account.md) |

### Администрирование {#admin}

| Цель | Страница |
|------|----------|
| Собрать и запустить сервер | [Установка и запуск](admin/install.md) |
| Настроить TLS, хранилище, федерацию, логирование, mailer / сброс пароля, captcha | [Конфигурация](admin/configuration.md) |

### Справка {#reference}

| Тема | Страница |
|------|----------|
| Сопоставление колонок ZIP Strava и поведение импорта | [Массовый импорт Strava](integrations/strava-bulk-import.md) |
| Импорт Health Sync из Google Drive (Android-клиент) | [Health Sync + Google Drive](integrations/health-sync-google-drive.md) |
| Полный аннотированный конфиг | [`config.full.yaml`](https://github.com/solargate/grom/blob/master/cmd/grom/config-examples/config.full.yaml) |
| HTTP API (Swagger UI) | На работающем сервере: `/api/docs/` (например `http://localhost:8080/api/docs/`); исходники OpenAPI в [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) |
| Политика конфиденциальности | [Privacy Policy](privacy.md) |
