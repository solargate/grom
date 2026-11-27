# Установка и запуск

Для быстрого запуска установите Grom из релиза GitHub. Веб-интерфейс встроен в бинарник сервера. Настройки для публичного сервера описаны в разделе [«Конфигурация»](configuration.md).

## Скачать с GitHub Releases {#download-from-github-releases}

Готовые пакеты публикуются с каждым [релизом GitHub](https://github.com/solargate/grom/releases):

| Артефакт | Содержимое |
|----------|------------|
| `grom-<version>-linux-amd64.tar.gz` | Бинарник сервера + `config-examples/` |
| `grom-<version>-darwin-arm64.zip` | Бинарник сервера + `config-examples/` (Apple Silicon) |
| `grom-<version>-darwin-amd64.zip` | Бинарник сервера + `config-examples/` (Intel Mac) |
| `grom-<version>-windows-amd64.zip` | Бинарник сервера + `config-examples/` |

В том же релизе доступен Android-клиент `grom-<version>.apk`.

## Запуск {#run}

1. Скачайте и распакуйте серверный архив для вашей ОС и архитектуры. Откройте терминал в распакованном каталоге, где находятся бинарник и `config-examples/`.
2. Скопируйте пример для локального HTTP в `config.yaml` и задайте в `auth.jwt_secret` длинный случайный секрет.
3. Запустите сервер с этим файлом.

**Linux и macOS**

```bash
cp config-examples/config.dev.notls.yaml config.yaml
# Перед запуском задайте auth.jwt_secret в config.yaml.
./grom --config config.yaml
```

**Windows PowerShell**

```powershell
Copy-Item config-examples/config.dev.notls.yaml config.yaml
# Перед запуском задайте auth.jwt_secret в config.yaml.
.\grom.exe --config config.yaml
```

Откройте `http://localhost:8080/` в браузере. Зарегистрируйтесь и войдите. Пример использует HTTP и предназначен для локальной проверки. Настройте [TLS](configuration.md#tls-profiles), прежде чем открывать доступ к серверу из интернета.

Без `--config` (или `-c`) Grom ищет `config.yaml` в текущем рабочем каталоге. Относительные пути хранилища отсчитываются от каталога бинарника.

Android-клиент может подключиться к тому же серверу, если устройство имеет к нему доступ. Укажите адрес сервера, доступный с этого устройства. Выбор адреса и вход описаны в [обзоре для пользователя](../user/overview.md). Swagger UI доступен по пути `/api/docs/` на работающем сервере.

## Дальше {#next-steps}

- Выберите профиль TLS и драйвер хранилища (`bbolt` для обычных установок; `file` в основном для тестов / крошечных инстансов) — [Конфигурация](configuration.md)
- Создайте резервную копию данных сервера и проверьте восстановление — [Резервное копирование и восстановление](backup-restore.md)
- При желании включите email для сброса пароля (`mailer` + `auth.reset.public_base_url`) — [Конфигурация — Mailer и сброс пароля](configuration.md#mailer-and-password-reset)
- При желании включите ALTCHA captcha на регистрацию/вход/«забыли пароль» (`auth.captcha.enabled`) — [Конфигурация — Auth captcha](configuration.md#auth-captcha-altcha)
- Сгенерируйте самоподписанные сертификаты для локального HTTPS — `grom gencerts` (см. раздел TLS в конфигурации)
- Обзор продукта для клиента — [Обзор для пользователя](../user/overview.md)
- Справка API в браузере — `/api/docs/` на работающем сервере

## Сборка из исходников {#build-from-source}

### Требования {#requirements}

- Go 1.26+
- Flutter SDK `>=3.4.0 <4.0.0` (для встроенного веб-UI и сборок Android)
- Python 3 (сайт документации MkDocs и tooling `server-catalog.yaml`; `make web` / `make android-*` пересобирают каталог)
- Make

### Сборка {#build}

Из корня репозитория:

```bash
make grom    # swagger + Flutter web + Go-бинарник → cmd/grom/grom
make test    # тесты Go и Flutter
make cli     # только Go-бинарник (без пересборки Flutter web)
make web     # Flutter web → internal/web/dist
```

Другие полезные цели: `make apidoc` (перегенерация OpenAPI), `make catalog` (Dart-каталог серверов), `make android-apk`, `make clean`.

После сборки перейдите в `cmd/grom` и выполните [шаги запуска](#run), используя бинарник и примеры конфигурации из этого каталога.
