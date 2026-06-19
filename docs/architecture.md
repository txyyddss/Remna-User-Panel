# Архитектура проекта

Репозиторий разделен по зонам ответственности рантайма:

```text
cmd/
  backend/             Go entrypoint для webhook/API и Mini App backend
  worker/              Go entrypoint фоновых задач
  migrate/             одноразовый Go entrypoint миграций

internal/
  app/                 wiring runtime-зависимостей
  config/              загрузка env-конфигурации
  db/                  pgx pool и schema_migrations
  httpapi/             net/http + chi маршруты
  i18n/                JSON locale catalog
  payments/            интерфейсы payment provider adapters
  telegram/            Telegram Bot API client
  webassets/           Mini App templates, themes, favicon/logo assets
  workers/             cancellable worker orchestration

pkg/plugin/            публичный Go SDK расширений
frontend/              Svelte/Vite Mini App и админка
deploy/                Dockerfile, nginx- и caddy-конфиги рантайма
data/                  данные рантайма, монтируемые в контейнеры
locales/               переводы бота и Web App
```

Основной `docker-compose.yml` находится в корне репозитория. Он собирает три
прикладных образа из `deploy/docker/Dockerfile`:

- `backend`: Go HTTP API, Telegram webhook, payment webhooks, Remnawave webhook,
  Mini App backend и проверка здоровья `/healthz`.
- `worker`: Go worker process для фоновых задач и очередей.
- `frontend`: статические Svelte-ассеты, которые отдает nginx.

Сервис `migrate` - одноразовый контейнер на базе backend-образа. Он применяет
Go migration chain через общую таблицу `schema_migrations`, затем `backend` и
`worker` стартуют только после успешного завершения `migrate`.

Основные команды:

```bash
docker compose up -d --build
docker compose run --rm migrate
docker compose logs -f backend worker frontend
npm run build:webapp
go test ./...
```
