# Импорт через Strava API (Android) {#strava-api-import-android}

На **Android** Grom может импортировать недавние тренировки напрямую через [Strava API](https://developers.strava.com/), используя **ваши** credentials API-приложения Strava (bring-your-own / BYO). На сервер Grom ничего не сохраняется: Client ID, Client Secret и OAuth-токены остаются на устройстве. В веб-клиенте этот поток **недоступен**.

Для полной истории (или активностей старше окна синхронизации) используйте [массовый импорт Strava](strava-bulk-import.md) (ZIP-архив).

## Требования {#prerequisites}

1. Аккаунт Strava с активной подпиской (Strava требует её для создания API-приложения).
2. Создайте API-приложение на [strava.com/settings/api](https://www.strava.com/settings/api).
3. Скопируйте **Client ID** и **Client Secret** в Grom.

Если Connect падает с HTTP **403** на обмене токена, проверьте, что Client ID и Client Secret совпадают с API-приложением, и повторите попытку.

Новые приложения Strava стартуют в single-player режиме (авторизоваться может только владелец app) — это подходит для личного BYO.

## Сценарий пользователя {#user-flow}

1. Откройте **Интеграция** → **Внешние сервисы** → **Strava** (только Android).
2. Включите **Импорт тренировок из Strava**.
3. Введите Client ID, Client Secret и при необходимости **Тренировок за sync** (по умолчанию **10**, максимум **200**).
4. Нажмите **Connect with Strava**, выдайте scope `activity:read` и убедитесь, что статус **OK**.
5. На **Главной** нажмите иконку синхронизации в app bar (как раньше у Health Sync).
6. Grom покажет диалог «Синхронизация…», затем snackbar с числом импортированных тренировок (или что новых нет).

Выключение переключателя только прячет кнопку на Главной; credentials и токены остаются. Выход из Grom тоже их не очищает.

## Правила синхронизации {#sync-rules}

| Правило | Поведение |
|---------|-----------|
| Лимит | До настраиваемого числа **Тренировок за sync** (по умолчанию **10**, максимум **200**) последних активностей |
| Порядок | От новых к старым |
| Стоп | На первой активности, уже есть в Grom с `external_id.name=strava` и тем же id |
| Видимость | Только `activity:read` — Everyone / Followers (не «Only You») |
| Без GPS | Workout по summary без трека |
| С GPS | GPX из streams Strava (включая пульс, если есть) |
| Устройство | Берёт Strava `device_name`, если есть (иначе дефолт сервера `Grom App`) |
| Фото | Best-effort через photos API; ошибка фото не валит workout |
| Снаряжение | `equipment_ids` не передаётся → сервер берёт `last_equipment_by_sport` |

Дедуп совпадает с ZIP-импортом (`strava` + id активности Strava).

## Конфиденциальность {#privacy-note}

Client secret и refresh-токены хранятся в настройках приложения на устройстве. Grom не отправляет их на ваш инстанс.

## Связанные страницы {#related}

- [Массовый импорт Strava](strava-bulk-import.md)
- [Импорт треков](import-tracks.md)
- [Обзор пользователя](../user/overview.md)
