---
name: solid-fiber
description: >-
  Guides full-stack development on the solid-fiber starter — a SolidJS + UnoCSS
  frontend and a Go + Fiber v3 + SQLite backend compiled into one static binary.
  Use this whenever working inside a solid-fiber project and adding or changing a
  backend resource/entity/endpoint, a database migration, a frontend API call or
  page/component, or when starting a new project from the template. Trigger it
  even when the user speaks loosely — "add a tags resource", "new endpoint for
  X", "let users comment on work items", "add a settings page", "scaffold a new
  app from this", "wire up the API for Y" — because the value of this stack is
  that every feature follows the same layered shape, and this skill encodes that
  shape plus the non-obvious gotchas that otherwise break the build.
---

# solid-fiber

A batteries-included full-stack starter. One repo, one deployable static binary:

- **Frontend** — SolidJS + Vite + UnoCSS (preset-wind4), built into `backend/web/dist` and embedded into the Go binary via `//go:embed`.
- **Backend** — Go + Fiber v3, SQLite (pure-Go `modernc.org/sqlite`, so the binary stays fully static), a layered architecture, and a standard JSON envelope.
- **Guardrails** — versioned migrations, structured logging, middleware (recover, request-id, helmet, rate limit, timeouts), a container healthcheck, full test + lint gates in CI, and Dependabot.

## The golden rule

**Mirror the nearest existing feature.** This stack's whole value is that every resource has the *same* shape, so an LLM (or a human) can add one by pattern-matching rather than inventing. The reference implementation is **`work_item`** — when adding anything backend, copy how `work_item` does it; when adding anything frontend, copy how `WorkItems`/`WorkItemCard` do it. Consistency here is a feature, not a constraint: it's what keeps the codebase legible and expandable.

## Project map

```
backend/
  main.go                     # config, DI wiring, graceful shutdown, --healthcheck mode
  api/
    server.go                 # app assembly: middleware + routes + error handler + SPA
    handlers/                 # HTTP handlers (thin; map errors -> status)
    presenter/                # JSON envelope shaping
    routes/                   # route registration per resource
  pkg/
    <resource>/               # one package per domain resource
      entities.go             # the struct
      repository.go           # Repository interface + in-memory impl
      sqlite_repository.go    # SQLite impl
      service.go              # use cases + validation (assigns UUID)
      migrations.go           # []storage.Migration for this resource
    storage/                  # OpenSQLite + the migration runner
  web/embed.go                # //go:embed of the built SPA
frontend/app/
  src/
    api.ts                    # typed fetch client (unwraps the envelope)
    <Feature>.tsx             # Solid components
    format.ts                 # shared pure UI helpers
    *.test.ts(x)              # Vitest
```

## Adding a full-stack feature — the workflow

Most requests ("add a X resource", "let users do Y") are a new resource or an extension of one. Work outside-in from the data model. Do the backend first (it defines the contract), then the frontend.

1. **Backend resource** — entity → migration → repository (interface + in-memory + SQLite) → service (validation + UUID) → handlers → routes → presenter → DI wiring in `api/server.go` and `main.go`. Full step-by-step with copy-ready templates: **read `references/backend-feature.md`**.
2. **Frontend** — a typed client function in `api.ts`, a component that uses `createResource`/`createSignal`, and Vitest coverage. Full templates: **read `references/frontend-feature.md`**.
3. **Verify before you're done** — run the gates in the "Verification gates" section below. Green gates are the definition of done.

For the conventions every layer relies on (the JSON envelope, error→status mapping, pagination, IDs, middleware): **read `references/conventions.md`**.

Keep `SKILL.md` open for the workflow and gates; open a reference file when you're actually writing that layer.

## Conventions cheat-sheet

These are load-bearing — the layers assume them. Details and rationale in `references/conventions.md`.

- **JSON envelope**: every response is `{"status": bool, "data": ..., "error": string|null}`; list endpoints add `"meta": {total, limit, offset}`. Presenter functions build these; handlers never hand-roll JSON.
- **IDs are server-assigned UUID strings.** The service sets `ID = uuid.NewString()` on create and ignores any client-supplied id.
- **Errors flow up as sentinel errors**, and `handlers.statusForError` maps them to HTTP codes (validation → 400, not-found → 404, else 500). Add new validation sentinels to that switch.
- **The Repository is an interface with two implementations** (in-memory for tests/dev, SQLite for real). Keep them in lockstep.
- **Migrations are append-only.** Add a new `storage.Migration` with the next version number; never edit an applied one.
- **Fiber v3, not v2**: body parsing is `c.Bind().Body(&x)` (there is no `c.BodyParser`); listen is `app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx})`.

## Verification gates (definition of done)

Run these from the repo root before declaring a feature complete. CI runs the same set, so green here means green there.

```sh
# Backend
cd backend && go build ./... && go vet ./... && go test ./...
"$(go env GOPATH)/bin/golangci-lint" run ./...   # see note below

# Frontend
cd frontend/app && bun install --frozen-lockfile && bun run lint && bun run typecheck && bun run test && bun run build
```

**Three gotchas that will bite you if you skip them** (each has burned a real change in this repo):

1. **Restore the embed placeholder after any frontend build.** `bun run build` uses Vite's `emptyOutDir`, which **deletes `backend/web/dist/.gitkeep`**. That file must stay committed or a fresh clone fails `//go:embed all:dist`. After building, always `touch backend/web/dist/.gitkeep` before staging/committing.
2. **golangci-lint must be built with the module's Go toolchain.** The prebuilt release refuses a module whose `go` directive is newer than its own build. Install with `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>` (the setup-go toolchain builds it correctly), matching the version CI pins.
3. **TypeScript is intentionally held at 5.x.** `typescript-eslint` doesn't support the TS 7 native compiler yet, so bumping `typescript` past 5.x crashes `bun run lint`. A Dependabot ignore keeps it pinned; leave it until the lint toolchain catches up.

If a gate fails, fix it before moving on — a half-wired feature that skips a layer (e.g. a handler with no route, or a repository method missing from one of the two impls) is the most common failure mode.

## Starting a new project from the template

When the user wants a fresh project rather than a feature, **read `references/new-project.md`** — it covers cloning, renaming the Go module and app, running `make dev` (Vite on :5173 proxying `/api` to Fiber on :3000), and the first build. The short version: it's `make install` then `make dev`, and the one rename that matters is the Go module path in `backend/go.mod` (and its imports).

## Reference files

- `references/backend-feature.md` — add/extend a backend resource, with templates for every layer.
- `references/frontend-feature.md` — add a frontend API call + component + tests.
- `references/conventions.md` — the envelope, error mapping, pagination, IDs, middleware, and why each exists.
- `references/new-project.md` — scaffold and run a new project from the template.
