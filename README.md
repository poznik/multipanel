# multipanel

Панель для агрегации метрик `telemt` с нескольких серверов.

Что уже есть:
- backend на `Go`;
- один TOML-конфиг;
- параллельный опрос нескольких `telemt` API;
- агрегированный snapshot;
- light debug UI без внешней frontend-сборки;
- Docker-сборка в один контейнер.

## Локальный запуск

```bash
cp config.example.toml config.toml
# отредактируйте точки telemt

go run . --config config.toml
```

Панель будет доступна на `http://127.0.0.1:8080`.

## Docker

```bash
docker compose up --build
```

По умолчанию контейнер читает `/app/config.toml`, который монтируется из локального `./config.toml`.

## Скачивание Готового Контейнера Из Git

Если нужен именно актуальный уже собранный production-контейнер из репозитория, используйте каталог `/opt/multipanel` и архив образа из `release/`.

1. Скачайте проект в `/opt/multipanel`:

```bash
sudo mkdir -p /opt/multipanel
sudo chown -R "$USER":"$USER" /opt/multipanel
git clone https://github.com/poznik/multipanel.git /opt/multipanel
cd /opt/multipanel
```

2. Загрузите готовый Docker image из архива:

```bash
docker load -i release/multipanel-prod-image.tar.gz
```

3. Подготовьте конфиг:

```bash
cp config.example.toml config.toml
```

Отредактируйте `config.toml` под ваши `telemt` endpoints.

4. Запустите контейнер без пересборки:

```bash
docker compose up -d --no-build
```

5. Проверьте сервис:

```bash
docker compose ps
curl http://127.0.0.1:8080/healthz
```

Важно:
- архив образа лежит в `release/multipanel-prod-image.tar.gz`;
- команда `docker load` создаёт локальный image с тегом `multipanel:prod`;
- для запуска именно загруженного образа используйте `docker compose up -d --no-build`.

## Разворачивание На Ubuntu

Предполагается, что `Docker` и `docker compose` уже установлены, а рабочая директория сервиса должна быть `/opt/multipanel`.

1. Подготовьте каталог:

```bash
sudo mkdir -p /opt/multipanel
sudo chown -R "$USER":"$USER" /opt/multipanel
cd /opt/multipanel
```

2. Перенесите файлы проекта в `/opt/multipanel`.

Если проект уже лежит в git:

```bash
git clone https://github.com/poznik/multipanel.git /opt/multipanel
cd /opt/multipanel
```

Если проект копируется вручную, достаточно положить в `/opt/multipanel` весь текущий каталог проекта.

3. Подготовьте конфиг:

```bash
cp config.example.toml config.toml
```

Отредактируйте `config.toml`: укажите реальные `telemt` endpoints, адрес прослушивания и при необходимости `Authorization` headers.

4. Соберите и запустите контейнер:

```bash
cd /opt/multipanel
docker compose up -d --build
```

5. Проверьте запуск:

```bash
docker compose ps
docker compose logs -f multipanel
curl http://127.0.0.1:8080/healthz
```

6. Управление сервисом:

Перезапуск после изменения `config.toml`:

```bash
docker compose restart multipanel
```

Остановка:

```bash
docker compose down
```

Обновление образа после изменения кода:

```bash
docker compose up -d --build
```

Важно:
- контейнер монтирует `/opt/multipanel/config.toml` в `/app/config.toml`;
- изменения конфига применяются после рестарта контейнера;
- compose-файл рассчитан на запуск из каталога `/opt/multipanel`;
- сервис публикуется на `0.0.0.0:8080`, если это не изменено в `config.toml`.

## Конфиг

```toml
[server]
listen = "0.0.0.0:8080"
refresh_interval = "10s"
request_timeout = "5s"

[telemt]
allow_insecure_tls = false

[[telemt.endpoints]]
name = "ams-1"
scheme = "http"
address = "telemt-1.example.net"
port = 2398
auth_header = ""
```

Поддерживаются:
- несколько endpoints;
- `http` и `https`;
- опциональный `Authorization` header для каждого endpoint;
- включение/отключение endpoint через `enabled = false`.

## Что Показывает UI

- суммарные метрики по всем серверам;
- таблицу по каждому серверу;
- `ME/CD mode`;
- `ME Quality` по DC: `RTT`, `Writers`, `Coverage`;
- статус доступности API и ошибки отдельных endpoint-запросов.

Примечание:
- `Total traffic` в текущей debug-версии считается по пользовательским octet counters из `/v1/users`.
