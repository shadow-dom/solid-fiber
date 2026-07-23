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

This is a multi-stage build: the SPA is built with bun, then embedded into a fully static Go binary, and the final image runs on a distroless nonroot base, listening on `:3000`.

## Project layout

```
backend/
  main.go              # entrypoint: dependency wiring, routes, embedded SPA + fallback, graceful shutdown
  api/
    handlers/           # HTTP handlers
    presenter/           # JSON response envelope shaping
    routes/              # route registration
  pkg/work_item/         # domain: entities, repository, service
  web/                    # //go:embed of the built SPA (backend/web/dist)
frontend/app/             # SolidJS application (Vite, UnoCSS)
```

## API

- `GET /api/hello` - health/sample endpoint.
- Work items CRUD under `/api/work-items`:
  - `POST /api/work-items` - create a work item.
  - `GET /api/work-items?project_id=<id>` - list work items for a project.
  - `GET /api/work-items/:id` - fetch a single work item.
  - `PUT /api/work-items/:id` - update a work item.
  - `DELETE /api/work-items/:id` - delete a work item.

All responses use a standard JSON envelope: `{"status": bool, "data": ..., "error": ...}`. Entity ids are UUID strings.
