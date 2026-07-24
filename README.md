# solid-fiber

A full-stack starter combining a [SolidJS](https://www.solidjs.com/) + [UnoCSS](https://unocss.dev/) (preset-wind4) frontend with a [Go](https://go.dev/) + [Fiber v3](https://github.com/gofiber/fiber) backend, compiled into a single static Go binary that embeds the built SPA via `//go:embed`.

## Quickstart

```sh
make install   # install frontend (bun) and backend (go) dependencies
make dev       # Vite HMR on :5173, proxying /api to Fiber on :3000
```

Open http://localhost:5173 during development.

```sh
make build     # builds the SPA into backend/web/dist, then the Go binary at bin/server
make run       # build + run the single embedded binary (serves everything on :3000)
```

## Docker

```sh
docker build -t solid-fiber .
docker run -p 3000:3000 solid-fiber
```

This is a multi-stage build: the SPA is built with bun, then embedded into a fully static Go binary, and the final image runs on a distroless nonroot base, listening on `:3000`. The image defines a `HEALTHCHECK` that runs the binary in `--healthcheck` mode (`server --healthcheck` probes `/api/health` and exits non-zero when unhealthy), so the shell-less runtime can self-report health.

Or use Docker Compose (`make up` / `make down`), which builds the image and mounts a named volume so the SQLite database survives restarts:

```sh
docker compose up --build -d
```

## Configuration

The server reads two environment variables:

| Variable  | Default            | Description                                              |
| --------- | ------------------ | -------------------------------------------------------- |
| `ADDR`    | `:3000`            | Address (host:port) the server binds to.                 |
| `DB_PATH` | `work_items.db`    | Path to the SQLite database file (created on first run). |

In the container image `DB_PATH` defaults to `/data/work_items.db`, and `/data` is a volume so the database persists. See `.env.example`.

## Persistence

Work items are stored in a SQLite database via the pure-Go [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver — no cgo, so the binary stays fully static. The schema is applied on startup by a small forward-only migration runner (`pkg/storage`), tracked in a `schema_migrations` table. The store is swappable: `pkg/work_item` defines a `Repository` interface with both a SQLite and an in-memory implementation.

## Project layout

```
backend/
  main.go              # entrypoint: config, dependency wiring, graceful shutdown
  api/
    server.go           # app assembly: middleware, routes, health, embedded SPA + fallback
    handlers/           # HTTP handlers (work items, health)
    presenter/           # JSON response envelope shaping
    routes/              # route registration
  pkg/work_item/         # domain: entities, service, repository (SQLite + in-memory), migrations
  pkg/storage/           # datastore connection + migration runner (SQLite)
  web/                    # //go:embed of the built SPA (backend/web/dist)
frontend/app/             # SolidJS application (Vite, UnoCSS)
```

## API

- `GET /api/hello` - sample endpoint.
- `GET /api/health` - liveness + datastore check (`200` healthy, `503` if the DB is unreachable).
- Work items CRUD under `/api/work-items`:
  - `POST /api/work-items` - create a work item.
  - `GET /api/work-items?project_id=<id>` - list work items for a project.
  - `GET /api/work-items/:id` - fetch a single work item.
  - `PUT /api/work-items/:id` - update a work item.
  - `DELETE /api/work-items/:id` - delete a work item.

All responses (including errors and 404s) use a standard JSON envelope: `{"status": bool, "data": ..., "error": ...}`. Entity ids are UUID strings.

Every request passes through middleware for panic recovery, a request id (`X-Request-Id`), security headers ([helmet](https://pkg.go.dev/github.com/gofiber/fiber/v3/middleware/helmet)), and structured (slog) access logging.
