# Conventions & why they exist

These hold the stack together. Follow them so features compose; the "why" is here
so you can adapt sensibly rather than cargo-cult.

## The JSON envelope

Every API response is:

```json
{ "status": true, "data": <payload>, "error": null }
```

and list endpoints add pagination metadata:

```json
{ "status": true, "data": [ ... ], "meta": { "total": 42, "limit": 20, "offset": 0 }, "error": null }
```

On failure: `{ "status": false, "data": null, "error": "message" }`. **Why:** a
single, predictable shape means the frontend has exactly one unwrap path
(`request<T>()` in `api.ts`) and one error path, and any client can rely on it.
Presenter functions (`api/presenter/`) are the only place this shape is built —
handlers call them, they never assemble JSON by hand.

## IDs

Server-assigned UUID strings. The service does `item.ID = uuid.NewString()` on
create and **ignores any client-supplied id**. **Why:** clients can't cause id
collisions or overwrite arbitrary rows, and the id is opaque/portable.

## Error handling → HTTP status

Domain layers return **sentinel errors** (`var ErrNotFound = errors.New(...)`,
`ErrTitleRequired`, …). Handlers translate them with a `statusForError` switch:

```go
switch {
case errors.Is(err, ...validation...): return http.StatusBadRequest   // 400
case errors.Is(err, work_item.ErrNotFound): return http.StatusNotFound // 404
default: return http.StatusInternalServerError                          // 500
}
```

The global Fiber `ErrorHandler` (in `api/server.go`) renders any *un*handled
error (unknown `/api/*` route, a recovered panic, a rate-limit hit) as the same
JSON envelope. **Why:** clients get consistent, typed failures instead of Fiber's
default plain-text.

## Pagination

List endpoints take `?limit=&offset=` (default limit 20, max 100, offset ≥ 0),
order by `id` for stability, and return `meta.total` from a `COUNT(*)`. **Why:**
bounded responses and a stable window as data grows. (There's no `created_at`
column yet, so "newest first" isn't possible without adding one — a known
follow-up.)

## Persistence & migrations

- SQLite via pure-Go `modernc.org/sqlite` (no cgo → the binary stays fully
  static and runs on distroless). Opened via `storage.OpenSQLite(path)` with WAL
  + busy-timeout + foreign-keys.
- Schema changes are **append-only** `storage.Migration` entries run by
  `storage.Migrate` (tracked in a `schema_migrations` table, each in its own
  transaction). Add a new version; never mutate an applied one. **Why:** every
  environment converges to the same schema deterministically on startup.
- The `Repository` is an **interface** with an in-memory impl (tests/dev) and a
  SQLite impl. **Why:** fast, dependency-free unit tests, and the store is
  swappable (a Postgres impl would just be a third implementation).

## Middleware (already wired in `api/server.go`)

Order matters: `requestid` → `requestLogger` (slog) → `recover` → `helmet`, then
per-group a `limiter` (100/min on `/api`, health exempt), plus server
`ReadTimeout`/`WriteTimeout`/`IdleTimeout` and a 1 MiB `BodyLimit`. New endpoints
inherit all of this for free — don't re-add it per handler.

## Config & runtime

- Env: `ADDR` (default `:3000`), `DB_PATH` (default `work_items.db`; `/data/work_items.db` in the image).
- `server --healthcheck` probes `/api/health` and exits 0/1 — the same static
  binary is the container `HEALTHCHECK` (distroless has no shell).
- Graceful shutdown via `fiber.ListenConfig{GracefulContext: ctx}` on SIGINT/SIGTERM.

## Fiber v3 differences from v2 (easy to get wrong)

- Body parsing: `c.Bind().Body(&x)` — **`c.BodyParser` does not exist in v3**.
- Params/query: `c.Params("id")`, `c.Query("limit")`.
- Listen with graceful context: `app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx})`.
- Middleware import for `recover` collides with the builtin; alias it
  (`recoverer "github.com/gofiber/fiber/v3/middleware/recover"`).

## Toolchain pins you must respect

- **golangci-lint** is built from source in CI with the module's Go toolchain,
  because the prebuilt binary refuses a module whose `go` directive is newer than
  its build. Locally, `go install .../golangci-lint/v2/cmd/golangci-lint@<pinned>`.
- **TypeScript is pinned to 5.x** (Dependabot ignores TS majors) because
  `typescript-eslint` doesn't support the TS 7 native compiler yet — TS 7 builds
  fine but crashes `bun run lint`.
- The frontend lockfile is the **text `bun.lock`** (not binary `bun.lockb`), so
  Dependabot can maintain it; the Dockerfile copies `bun.lock*`.

## The embed placeholder (the #1 footgun)

`backend/web/dist/.gitkeep` is committed so `//go:embed all:dist` resolves in a
fresh clone. Vite's `emptyOutDir` **deletes it on every `bun run build`**. Always
`touch backend/web/dist/.gitkeep` after building and before committing. Built
assets (`index.html`, `assets/*`) are gitignored; only the placeholder is tracked.
