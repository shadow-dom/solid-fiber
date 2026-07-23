# solid-fiber — Code Review & Handoff Task List

**Reviewed commit:** `37defa0` (branch `claude/pr-task-list-handoff-kfn7fj`, identical to `master`)
**Reviewer:** Claude Code
**Date:** 2026-07-23

> There is no open pull request in this repo, so this is a **full review of the current tree**, not a diff review. Tasks are granular, ordered by severity, and each is tagged with a suggested model:
> - **[Sonnet]** — well-scoped, low-ambiguity, mechanical or single-file fixes with a clear correct answer.
> - **[Opus 4.8]** — cross-cutting, architectural, or judgment-heavy work touching multiple files or requiring design decisions.

## Build status (as found)

The project **did not build**. `go build ./...` failed for *three independent reasons* (a third surfaced during the fix), plus a broken clone:

1. `//go:embed all:dist` failed: `pattern all:dist: no matching files found` (no `dist/` dir was tracked).
2. `backend/api/handlers` referenced an undefined `presenter` package.
3. **(discovered during fix)** `main.go` lived in `backend/cmd/`, but the Makefile (`go run .`, `go build … .`) and Dockerfile (`go build … .`) all build the module root `backend/`, which had no Go files — so the backend never built via its own tooling. Also, the handler used `c.BodyParser(...)`, which Fiber v3 removed (now `c.Bind().Body(...)`).

## ✅ Completed on branch `claude/pr-task-list-handoff-kfn7fj`

The P0 blockers and the core (interdependent, Opus-level) feature work are **done, building, vetted, and tested** in this branch. Decision: **UUID string IDs**.

- **T1** — committed `backend/web/dist/.gitkeep`; embed resolves.
- **T2** — added `backend/api/presenter/work_item.go` (`{status, data, error}` envelope).
- **T3** — removed the dangling `examples/go-project` gitlink.
- **T4** — correct status codes (400 empty title, 404 not found, 201 created, 204 delete) via `statusForError`.
- **T5** — `WorkItem.ID` and all FK fields are now UUID `string`; service assigns a server-side UUID on create.
- **T6** — full feature wired: `Repository` (in-memory impl) → `Service` → CRUD handlers → `routes.WorkItemRouter` → DI in `main.go`.
- **T7** — unmatched `/api/*` now returns a JSON 404, not the SPA shell.
- **T8** — graceful shutdown via `fiber.ListenConfig{GracefulContext}` on SIGINT/SIGTERM.
- **T9** — table-driven service tests + handler integration tests (`go test ./...` green).
- **Discovered blocker #3** — moved `main.go` to `backend/` root; migrated `BodyParser` → `Bind().Body`.

**Verification:** `go build ./...`, `go vet ./...`, `go test ./...` all pass; frontend `bun run build` succeeds.

### ✅ Second wave — completed (all remaining tasks)

Dispatched as three parallel subagents (disjoint file ownership), then integrated and re-verified together.

- **T10** [Sonnet] — `main.go` now uses `log/slog` (JSON handler); fatal paths log + `os.Exit(1)`.
- **T11** [Sonnet] — deleted `pnpm-lock.yaml`; Dockerfile copies only `bun.lockb`. bun is the single lockfile.
- **T12** [Sonnet] — `package.json` renamed `solid-fiber` with a real description.
- **T13** [Sonnet] — `typecheck` script added; `build` now runs `tsc --noEmit` first (`--skipLibCheck` to avoid `node_modules` `.d.ts` noise from config files not in `include`).
- **T14** [Opus] — `.github/workflows/ci.yml`: backend (build/vet/test) + frontend (install/typecheck/build) jobs.
- **T15** [Sonnet] — corrected the "OKLCH" comment to HSL.
- **T16** [Sonnet] — real `<title>`; light/dark `theme-color` metas.
- **T17** [Opus] — `useTheme()` refactored: one-time `init()` (guarded flag + `typeof window`), persistence effect owned by a `createRoot`; accessors-only return. No more per-call listener/effect duplication.
- **T18** [Sonnet] — icon renders via `<Dynamic component={meta.Icon} />`.
- **T19** [Sonnet] — root `README.md` added.
- **T20** [Sonnet] — frontend `README.md` rewritten (bun, :5173 proxy, scripts, embed output).

**Verification (combined):** `go build/vet/test` green; `bun install --frozen-lockfile` (no changes), `bun run typecheck`, `bun run build` all pass.

**All 20 review tasks + 1 discovered blocker are complete.** Detailed original findings remain below for reference.

---

## P0 — Blockers (repo does not build / clone cleanly)

### T1 — [Sonnet] Restore the embedded `dist/` directory so `//go:embed` resolves
- **Where:** `backend/web/embed.go:8` (`//go:embed all:dist`); `backend/web/dist/` is absent from git.
- **Problem:** `go build` fails with `pattern all:dist: no matching files found`. The `Makefile clean` target and `.dockerignore` both reference `backend/web/dist/.gitkeep`, but that placeholder was never committed, so the directory doesn't exist in a fresh clone.
- **Fix:** Commit `backend/web/dist/.gitkeep` (the `all:` prefix will embed the dotfile so the pattern matches even when the SPA hasn't been built). Confirm `.gitignore`/`.dockerignore` keep the placeholder (`!backend/web/dist/.gitkeep`) while ignoring real build output — they already have the negation, so just add the file.
- **Verify:** `cd backend && go build ./...` gets past the embed error.

### T2 — [Opus 4.8] Fix the `handlers` package: the `presenter` layer is referenced but does not exist
- **Where:** `backend/api/handlers/work_item_handler.go:17,21,27,29` call `presenter.WorkItemErrorResponse(...)` and `presenter.WorkItemSuccessResponse(...)`; there is no `presenter` package and no import for it anywhere in the module.
- **Problem:** Compile error (`undefined: presenter`). The response-shaping layer was never written.
- **Fix:** Decide the API envelope and create `backend/api/presenter` (e.g. `presenter.WorkItemSuccessResponse(item *work_item.WorkItem) fiber.Map` and `presenter.WorkItemErrorResponse(err error) fiber.Map`), then import it in the handler. This is tagged Opus because the response contract (error shape, status semantics, success envelope) is a design decision that should be consistent across all future handlers, not a mechanical stub.
- **Verify:** `go build ./...` and `go vet ./...` are clean.

### T3 — [Sonnet] Remove or properly declare the dangling `examples/go-project` submodule
- **Where:** `examples/go-project` is committed as a gitlink (mode `160000`, commit `f5ba7177aecc56e99ad0d81db7b4311736fb8feb`), but there is **no `.gitmodules` file**.
- **Problem:** A gitlink with no submodule config is a broken reference: `git submodule update --init` can't resolve a URL, and the directory clones empty. It also makes `find examples -type f` return nothing.
- **Fix:** Either (a) remove the gitlink entirely — `git rm --cached examples/go-project` and delete the empty dir — if the example isn't ready, or (b) add a real `.gitmodules` entry with a valid URL if it's meant to be a submodule. Given there's no upstream config, (a) is almost certainly correct. Confirm intent with the repo owner if unsure.

---

## P1 — Correctness bugs

### T4 — [Sonnet] Wrong HTTP status codes in `AddWorkItem`
- **Where:** `backend/api/handlers/work_item_handler.go:20` and `:26`.
- **Problem:** A missing `Title` (client input error) returns `500 Internal Server Error`; it should be `400 Bad Request`. Likewise, `service.CreateWorkItem` failures are unconditionally `500`, collapsing validation/conflict errors into server errors.
- **Fix:** Return `http.StatusBadRequest` for the empty-title check. For the service error, either map known error types to appropriate 4xx codes or leave `500` only for genuinely unexpected errors. Straightforward once T2's error envelope exists.

### T5 — [Opus 4.8] `WorkItem.ID` type is inconsistent with the rest of the model and the service interface
- **Where:** `backend/pkg/work_item/entities.go:4` (`ID string`), vs `service.go:5,7` (`GetWorkItemByID(id int)`, `DeleteWorkItem(id int)`) and the `int` foreign keys (`ParentID`, `AssigneeID`, etc.).
- **Problem:** The primary key is a `string` while every reference to a work item ID elsewhere is an `int`. This will bite as soon as persistence and lookups are wired up.
- **Fix:** Pick one representation for IDs across entities and the service interface (int auto-increment vs string/UUID — note `google/uuid` is already an indirect dep, hinting at UUID intent) and apply it consistently to `ID`, `GetWorkItemByID`, `DeleteWorkItem`, and the FK fields. Architectural because it defines the data model's identity strategy.

### T6 — [Opus 4.8] The entire `work_item` feature is defined but never wired up
- **Where:** `backend/cmd/main.go` only registers `GET /api/hello` and the SPA fallback. `work_item.Service` has no concrete implementation, and `handlers.AddWorkItem` is never mounted on a route.
- **Problem:** `Service` is an interface with no implementor; `AddWorkItem`, the entity, and the interface are dead code. There's no persistence layer, no route wiring, and no dependency injection.
- **Fix:** Implement `Service` (start with an in-memory repo, or a real store), register CRUD routes under `/api/work-items` in `main.go`, and inject the service into the handlers. Depends on T2 and T5. This is the core feature build-out — Opus.

### T7 — [Sonnet] SPA fallback swallows unknown `/api/*` routes as `200 index.html`
- **Where:** `backend/cmd/main.go:33` (`app.Get("/*")`).
- **Problem:** A request to an unregistered API path (e.g. `GET /api/does-not-exist`) falls through to the catch-all and returns `index.html` with `200`, instead of a `404` JSON error. Confusing for API clients.
- **Fix:** Exclude the `/api` prefix from the SPA fallback (return `fiber.ErrNotFound` when `c.Path()` starts with `/api`), or register a `404` handler on the `api` group.

---

## P2 — Robustness & architecture

### T8 — [Opus 4.8] No graceful shutdown
- **Where:** `backend/cmd/main.go:46` (`log.Fatal(app.Listen(addr))`).
- **Problem:** The server exits hard on signal; in-flight requests are dropped and there's no cleanup hook (important once a DB/connection pool exists).
- **Fix:** Listen in a goroutine, catch `SIGINT`/`SIGTERM` with `signal.NotifyContext`, and call `app.Shutdown()` (or `ShutdownWithTimeout`) on shutdown. Pairs naturally with T6's resource wiring.

### T9 — [Opus 4.8] No test coverage anywhere
- **Where:** whole repo — no `*_test.go`, no frontend test setup.
- **Problem:** Nothing guards the handler validation, service behavior, or the theme logic.
- **Fix:** Add table-driven tests for `AddWorkItem` (bad body, empty title, success) and for the `Service` implementation once it exists; add a minimal Vitest setup for `theme.ts` (mode resolution, palette persistence). Tagged Opus because it also means choosing the test tooling/conventions for the project.

### T10 — [Sonnet] Structured logging instead of `log.Fatalf`
- **Where:** `backend/cmd/main.go:24,46`.
- **Problem:** Uses the stdlib `log` package with no levels/structure. Fine for a toy, worth upgrading for anything real.
- **Fix:** Switch to `log/slog` with a JSON handler; keep it small. Low priority.

---

## P3 — Build tooling & config

### T11 — [Sonnet] Two competing lockfiles committed (`bun.lockb` + `pnpm-lock.yaml`)
- **Where:** `frontend/app/bun.lockb` and `frontend/app/pnpm-lock.yaml`.
- **Problem:** The Makefile and Dockerfile use **bun** (`bun install --frozen-lockfile`), but a `pnpm-lock.yaml` (a Solid-template leftover) is also committed. Mixed lockfiles drift and confuse `--frozen-lockfile`.
- **Fix:** Pick bun (matches the tooling) and delete `pnpm-lock.yaml`, or vice versa. Update the Dockerfile `COPY` glob accordingly so it only pulls the chosen lockfile.

### T12 — [Sonnet] `package.json` still carries template identity
- **Where:** `frontend/app/package.json:2-4`.
- **Problem:** `"name": "vite-template-solid"`, `"version": "0.0.0"`, empty `"description"`.
- **Fix:** Rename to `solid-fiber` (or the app's real name), set a description. Trivial.

### T13 — [Sonnet] No typecheck step in the frontend build/CI
- **Where:** `frontend/app/package.json` scripts; `vite build` does not run `tsc`.
- **Problem:** Type errors won't fail the build. `tsconfig` has `noEmit: true`, so `tsc` is purely a checker but isn't invoked.
- **Fix:** Add a `"typecheck": "tsc --noEmit"` script and wire it into `build` (`"build": "tsc --noEmit && vite build"`) or a CI step.

### T14 — [Opus 4.8] No CI workflow
- **Where:** repo root — no `.github/workflows/`.
- **Problem:** Nothing enforces build/test/lint on push. The `.dockerignore` even lists `.github` as "CI noise", implying one was intended.
- **Fix:** Add a GitHub Actions workflow: Go build + `go vet` + tests, and frontend install + typecheck + build. Tagged Opus for the matrix/caching design and because it should gate the P0/P1 fixes above.

---

## P4 — Frontend polish

### T15 — [Sonnet] Misleading comment: "OKLCH" values are actually HSL
- **Where:** `frontend/app/src/ThemeToggle.tsx:41` — comment says "uses raw OKLCH values that match theme.css", but `PALETTE_SWATCH` values (and `theme.css`) are `hsl(...)`.
- **Fix:** Change the comment to say HSL. One line.

### T16 — [Sonnet] Default `index.html` metadata
- **Where:** `frontend/app/index.html:7` (`<title>Solid App</title>`) and `:6` (`theme-color` hardcoded `#000000`).
- **Problem:** Template default title; the `theme-color` is fixed black regardless of light/dark mode.
- **Fix:** Set a real `<title>` (e.g. "solid-fiber"). Optionally provide two `theme-color` metas with `media="(prefers-color-scheme: ...)"` so mobile browser chrome matches the theme.

### T17 — [Opus 4.8] `useTheme()` registers global side effects on every call
- **Where:** `frontend/app/src/theme.ts:33-64`.
- **Problem:** `mode`/`palette`/`resolved` are module-level singletons, but `useTheme()` sets up an `onMount` listener and a `createEffect` **each time it's called**. It's used once today (`ThemeToggle`), so it works — but a second consumer would double-register the `matchMedia` listener and the persistence effect.
- **Fix:** Move the one-time setup (media listener + persistence effect) into a module-level init that runs once (or a proper context provider), and have `useTheme()` only return the accessors/setters. Tagged Opus because it's a small architectural refactor (singleton store vs context provider) with reactivity implications.

### T18 — [Sonnet] Non-idiomatic component invocation `meta.Icon({})`
- **Where:** `frontend/app/src/ThemeToggle.tsx:136` (`{(meta.Icon({}) as JSX.Element)}`).
- **Problem:** Calls the component as a plain function and casts, bypassing Solid's normal `<Component />` handling.
- **Fix:** Render via `<Dynamic component={meta.Icon} />` (from `solid-js/web`) or `<meta.Icon />`. Low risk, improves readability.

---

## P5 — Documentation & housekeeping

### T19 — [Sonnet] No root README
- **Where:** repo root (only `frontend/app/README.md` exists, and it's a Solid-template leftover).
- **Fix:** Add a root `README.md`: what the stack is (Solid + UnoCSS preset-wind4 + Fiber, single embedded binary), how to `make dev` / `make build` / `make run`, and the Docker build. The `.dockerignore` already excludes `README.md` from the image, so this is docs-only.

### T20 — [Sonnet] Replace or remove the template `frontend/app/README.md`
- **Where:** `frontend/app/README.md`.
- **Problem:** It's the generic Solid Vite template readme — mentions pnpm, says "this file can be safely removed once you clone a template", and points to `http://localhost:3000` (wrong; dev is Vite on `:5173`).
- **Fix:** Delete it or replace with app-specific frontend notes.

---

## Suggested triage / ordering

| Order | Tasks | Rationale |
|------|-------|-----------|
| 1 | T1, T3 | Make the repo build & clone cleanly (mechanical). |
| 2 | T2, T5, T6 | Design the API/data layer and build the feature (Opus, sequential). |
| 3 | T4, T7 | Handler correctness, depends on T2. |
| 4 | T8, T9, T14 | Robustness, tests, CI. |
| 5 | T10–T13, T15–T20 | Tooling, polish, docs (mostly parallelizable Sonnet work). |

**Model split:** Sonnet → T1, T3, T4, T7, T10, T11, T12, T13, T15, T16, T18, T19, T20. Opus 4.8 → T2, T5, T6, T8, T9, T14, T17.
