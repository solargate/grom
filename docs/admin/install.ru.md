# Установка и запуск

Установите Grom из GitHub Release или соберите из исходников. Подробности конфигурации (TLS, хранилище, федерация, mailer / сброс пароля, captcha) — в [Конфигурация](configuration.md).

## Скачать с GitHub Releases {#download-from-github-releases}

Готовые пакеты публикуются с каждым [релизом GitHub](https://github.com/solargate/grom/releases):

| Артефакт | Содержимое |
|----------|------------|
| `grom-<version>-linux-amd64.tar.gz` | Бинарник сервера + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Бинарник сервера + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Бинарник сервера + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Бинарник сервера + `config-examples/` |
| `grom-<version>.apk` | Android-клиент |

Распакуйте архив для вашей ОС, скопируйте конфиг из `config-examples/` (задайте `auth.jwt_secret`), затем запустите бинарник `grom` — см. [Запуск](#run) ниже.

## Сборка из исходников {#build-from-source}

### Требования {#requirements}

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (для встроенного веб-UI и сборок Android)
- Make

### Сборка {#build}

Из корня репозитория:

```bash
make grom    # swagger + Flutter web + Go-бинарник → cmd/grom/grom
make test    # тесты Go и Flutter
make cli     # только Go-бинарник (без пересборки Flutter web)
make web     # Flutter web → internal/web/dist
```

Другие полезные цели: `make doc` (перегенерация OpenAPI), `make android-apk`, `make clean`.

## Запуск {#run}

Примеры конфигов есть в релизных архивах и в `cmd/grom/config-examples/` в репозитории. Сервер ищет `config.yaml` в текущем рабочем каталоге, если не передан `--config` / `-c`.

**Из релизного пакета** (пример только HTTP):

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# отредактируйте config.yaml — задайте auth.jwt_secret
./grom --config config.yaml
```

**Из исходников, только HTTP** (без TLS, федерация выключена):

```bash
cd cmd/grom && go run . --config config-examples/config.dev.notls.yaml
```

**После `make grom`**, с конфигом рядом с бинарником:

```bash
cd cmd/grom
./grom --config config.yaml
# или полагаться на config.yaml в рабочей директории по умолчанию:
./grom
```

Откройте **веб-UI** в браузере по базовому URL сервера (тот же Flutter-клиент, что на Android — например `http://localhost:8080/` с `config.dev.notls.yaml`). Зарегистрируйте пользователя и войдите. См. [Обзор для пользователя](../user/overview.md).

Приложение **Android** (и позже iOS) может подключиться к тому же инстансу: на экране входа укажите хост (схема необязательна). Открытый **HTTP разрешён для локальных/LAN**-установок без TLS; используйте HTTPS, если сервер доступен за пределами локальной сети. Как клиент выбирает `http` vs `https` — в [Обзор для пользователя](../user/overview.md).

**Документация API (Swagger UI):** `http://<host>:<port>/api/docs/` (например `http://localhost:8080/api/docs/`). Сгенерированные исходники OpenAPI также лежат в [`api/docs/`](https://github.com/solargate/grom/tree/master/api/docs) в репозитории.

Справка CLI:

```bash
grom --help
grom --version
```

## Дальше {#next-steps}

- Выберите профиль TLS и драйвер хранилища (`bbolt` для обычных установок; `file` в основном для тестов / крошечных инстансов) — [Конфигурация](configuration.md)
- При желании включите email для сброса пароля (`mailer` + `auth.reset.public_base_url`) — [Конфигурация — Mailer и сброс пароля](configuration.md#mailer-and-password-reset)
- При желании включите ALTCHA captcha на регистрацию/вход/«забыли пароль» (`auth.captcha.enabled`) — [Конфигурация — Auth captcha](configuration.md#auth-captcha-altcha)
- Сгенерируйте самоподписанные сертификаты для локального HTTPS — `grom gencerts` (см. раздел TLS в конфигурации)
- Обзор продукта для клиента — [Обзор для пользователя](../user/overview.md)
- Справка API в браузере — `/api/docs/` на работающем сервере
