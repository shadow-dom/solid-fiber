---
name: solid-fiber
description: >-
  Guides full-stack development on the solid-fiber starter — a SolidJS + UnoCSS
  frontend (shadcn/ui design language) and a Go + Fiber v3 + SQLite backend
  compiled into one static binary. Use this whenever working inside a solid-fiber
  project and adding or changing a backend resource/entity/endpoint, a database
  migration, a frontend API call or page/component, or when starting a new
  project from the template. Trigger it even when the user speaks loosely — "add
  a tags resource", "new endpoint for X", "let users comment on work items", "add
  a settings page", "scaffold a new app from this", "wire up the API for Y" —
  because the value of this stack is that every feature follows the same layered
  shape, built test-first, and this skill encodes that shape, the non-obvious
  gotchas, and the project's coding standards.
---

# solid-fiber

A batteries-included full-stack starter. One repo, one deployable static binary:

- **Frontend** — SolidJS + Vite + UnoCSS (preset-wind4) with a shadcn/ui-derived design system, built into `backend/web/dist` and embedded into the Go binary via `//go:embed`.
- **Backend** — Go + Fiber v3, SQLite (pure-Go `modernc.org/sqlite`, so the binary stays fully static), a layered architecture, and a standard JSON envelope.
- **Guardrails** — versioned migrations, structured logging, middleware (recover, request-id, helmet, rate limit, timeouts), a container healthcheck, full test + lint gates in CI, and Dependabot.

## Core principles

These are the standards this project is built to. They are not negotiable polish — they are what keep the codebase expandable, so apply them from the first line.

### 1. Let the code explain itself

Write code whose names and structure make its intent obvious, so comments are rarely needed. **A comment that explains _what_ the code does is a smell** — it means the code isn't clear enough. Fix the code, don't annotate it: rename the variable, extract a well-named function, simplify the branch. If a block seems to need a paragraph of explanation to be understood, that is a signal the block is wrong and should be rewritten or thrown out, not documented.

The only comments worth keeping are short notes on _why_ something non-obvious is done (a subtle ordering requirement, a workaround for an upstream bug) — and even those should be a line, not an essay. Doc comments on exported Go identifiers stay (they're API surface), but keep them to a crisp sentence. When you touch existing code that violates this, prefer to clean it up.

### 2. Test-driven development

Write the test first, watch it fail, then write the code that makes it pass, then refactor. This isn't ceremony — it forces you to state the behavior as a contract before you implement it, keeps everything testable by construction, and means a finished feature is a green feature. The layered architecture exists to make this cheap: the `Repository` interface + in-memory implementation let you test the service with no database, and `app.Test` drives handlers with no network. **Do not write a layer before its test.** The workflow below is ordered red→green→refactor for exactly this reason.

### 3. Mirror the reference implementation

Every feature has the _same_ shape, so add one by pattern-matching rather than inventing. The reference resource is **`work_item`** — copy how it does each layer (backend) and how `WorkItems`/`WorkItemCard` do it (frontend). Consistency here is the feature: it's what lets the next person (or model) extend the app without re-deriving the design.

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
    <resource>/               # one package per domain resource (entity, repo, service, migrations)
    storage/                  # OpenSQLite + the migration runner
  web/embed.go                # //go:embed of the built SPA
frontend/app/src/
  components/ui/              # shadcn/ui-style primitives (Button, Input, Card, …)
  api.ts                      # typed fetch client (unwraps the envelope)
  <Feature>.tsx              # Solid feature components (compose from components/ui)
  *.test.ts(x)               # Vitest
```

## Adding a full-stack feature — the workflow (test-first)

Most requests ("add a X resource", "let users do Y") are a new resource or an extension of one. Work outside-in from the data model, and inside each layer go red→green→refactor.

**Backend** (contract → tests → implementation, per layer):
1. Define the entity and the `Repository`/`Service` interfaces — just enough to compile.
2. **Write the service tests first** (validation, create-assigns-UUID, not-found), watch them fail, then implement the service + in-memory repository to pass.
3. **Write the SQLite repository test first** (CRUD, persist-across-reopen), then implement the SQLite impl + migration to pass.
4. **Write the handler tests first** (status codes, envelope, pagination via `app.Test`), then implement handlers + presenter + routes, and wire into `api/server.go` + `main.go`.

Full step-by-step with copy-ready templates: **read `references/backend-feature.md`**.

**Frontend** (test → implementation):
5. **Write the component/api test first** (renders items from a mocked `fetch`, create posts the right body), then implement the `api.ts` client + the component. Compose the UI from `components/ui` primitives — **read `references/ui.md`** for the shadcn/ui design language, and `references/frontend-feature.md` for the Solid patterns and templates.

6. **Verify** — run the gates in "Verification gates" below. Green gates are the definition of done.

For the conventions every layer relies on (envelope, error→status mapping, pagination, IDs, middleware): **read `references/conventions.md`**.

## Conventions cheat-sheet

Load-bearing; details and rationale in `references/conventions.md`.

- **JSON envelope**: every response is `{"status": bool, "data": ..., "error": string|null}`; list endpoints add `"meta": {total, limit, offset}`. Presenters build these; handlers never hand-roll JSON.
- **IDs are server-assigned UUID strings.** The service sets `ID = uuid.NewString()` on create and ignores any client id.
- **Errors flow up as sentinel errors**; `handlers.statusForError` maps them (validation → 400, not-found → 404, else 500). Add new validation sentinels there.
- **The Repository is an interface with two implementations** (in-memory for tests/dev, SQLite for real). Keep them in lockstep.
- **Migrations are append-only and namespaced per resource.** `storage.Migrate(db, "<table>", Migrations)` scopes versions by namespace, so each resource numbers its migrations from `1` with no collision. Never edit an applied one.
- **Fiber v3, not v2**: body parsing is `c.Bind().Body(&x)` (no `c.BodyParser`); listen is `app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx})`.

## UI: shadcn/ui design language

The theme tokens in `frontend/app/src/styles/theme.css` _are_ shadcn/ui's (background/foreground/card/primary/muted/accent/destructive/border/ring + radius, light/dark + palettes). Build UI the shadcn/ui way: **compose accessible, variant-driven primitives from `src/components/ui/` rather than inlining ad-hoc class strings.** Full guidance (primitives to build, Kobalte for accessible behavior, the `cn()` helper, a11y expectations): **read `references/ui.md`**.

## Verification gates (definition of done)

Run from the repo root before declaring a feature complete. CI runs the same set.

```sh
# Backend
cd backend && go build ./... && go vet ./... && go test ./...
"$(go env GOPATH)/bin/golangci-lint" run ./...

# Frontend
cd frontend/app && bun install --frozen-lockfile && bun run lint && bun run typecheck && bun run test && bun run build
```

**Three gotchas that will bite you if you skip them:**

1. **Restore the embed placeholder after any frontend build.** `bun run build` uses Vite's `emptyOutDir`, which deletes `backend/web/dist/.gitkeep`. That file must stay committed or a fresh clone fails `//go:embed all:dist`. After building, `touch backend/web/dist/.gitkeep` before staging.
2. **golangci-lint must be built with the module's Go toolchain** (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<pinned>`); the prebuilt release refuses a module whose `go` directive is newer than its build.
3. **TypeScript is held at 5.x** (Dependabot ignores TS majors) because `typescript-eslint` doesn't support the TS 7 native compiler yet — bumping it crashes `bun run lint`.

## Starting a new project from the template

When the user wants a fresh project rather than a feature, **read `references/new-project.md`**. Short version: `make install` then `make dev`, and the one rename that matters is the Go module path in `backend/go.mod` and its imports.

## Reference files

- `references/backend-feature.md` — add/extend a backend resource, test-first, with templates for every layer.
- `references/frontend-feature.md` — add a frontend API call + component + tests.
- `references/ui.md` — the shadcn/ui design language: primitives, Kobalte, accessibility.
- `references/conventions.md` — the envelope, error mapping, pagination, IDs, middleware, and why each exists.
- `references/new-project.md` — scaffold and run a new project from the template.
