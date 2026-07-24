# Starting a new project from the template

The repo *is* the template — start a project by copying it and renaming the Go
module. There's no generator CLI; that's deliberate (fewer moving parts, and the
LLM can adapt anything).

## 1. Get the code

```sh
# Use it as a GitHub template, or:
git clone <this-repo> my-app && cd my-app
rm -rf .git && git init      # start fresh history (optional)
```

## 2. Rename the Go module (the one rename that matters)

The module path `github.com/shadow-dom/solid-fiber` appears in `backend/go.mod`
and in every backend import. Replace it consistently:

```sh
cd backend
NEW=github.com/you/my-app
sed -i "s#github.com/shadow-dom/solid-fiber#$NEW#g" go.mod $(grep -rl shadow-dom/solid-fiber --include='*.go' .)
go build ./...   # confirm imports resolve
```

## 3. Rebrand the surface (optional but quick)

- `frontend/app/package.json` → `name`, `description`.
- `frontend/app/index.html` → `<title>`.
- `frontend/app/src/App.tsx` → header text.
- `README.md`, `LICENSE` copyright.

## 4. Install and run

```sh
make install                 # bun install + go mod tidy
make dev                     # Vite (HMR) on :5173 proxying /api -> Fiber on :3000
# open http://localhost:5173
```

`make build` produces the single embedded binary at `bin/server`; `make run`
builds and runs it; `make up`/`make down` are Docker Compose; `make test` and
`make lint` run the gates.

## 5. Make it yours

The template ships one example resource, `work_item`, end to end (backend +
UI + tests). Two paths from here:

- **Keep and extend it** — rename `work_item` to your first real resource, or add
  new resources beside it (see `references/backend-feature.md`).
- **Start clean** — delete the `work_item` package, its handlers/routes/presenter,
  the `WorkItems*`/`format` frontend files and their tests, and remove the
  wiring in `main.go`/`api/server.go`. Keep `pkg/storage`, the middleware in
  `api/server.go`, the health endpoint, and the theme system — that's the
  reusable spine.

## What you get out of the box

Persistence (SQLite + migrations), the JSON-envelope API with pagination and
validation, middleware (recover, request-id, helmet, rate limit, timeouts),
structured logging, a container healthcheck, a themed SolidJS UI with light/dark
+ palettes, full test + lint gates, CI, and Dependabot. Build features by
mirroring `work_item`; verify with the gates in `SKILL.md`.
