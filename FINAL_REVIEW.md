# solid-fiber — Final Review & Task List

**Scope:** a full, fresh-eyed review of the repository after it grew from a broken scaffold into a working full-stack starter. This supersedes the original `REVIEW_TASKS.md` (whose tasks are all done).

**Overall:** the stack is healthy — SQLite persistence + versioned migrations, hardened middleware (recover, request-id, helmet, structured logs, timeouts), a container healthcheck, full CRUD + pagination, a rich UI, backend + frontend test suites, lint gates in CI, and Dependabot. **No build-breakers or critical bugs remain.** Everything below is productionization and polish.

Tags: **[Sonnet]** = mechanical / well-scoped · **[Opus]** = design / cross-cutting.

## A. Security & tenancy
- **A1 — No authentication/authorization** `[Opus]`. The API is fully open and work items have no owner/tenant; anyone can read or mutate any project's items. The main blocker before real use. Add auth (session/JWT/API-key) + an ownership column and enforcement. *High effort.*
- **A2 — DB calls ignore request context** `[Sonnet→Opus]`. `pkg/work_item/sqlite_repository.go` uses `db.Query`/`Exec`, not `…Context`. Thread `c.Context()` through so cancelled/timed-out requests cancel DB work.
- **A3 — No rate limiting** `[Sonnet]`. Add Fiber's `limiter` middleware in `api/server.go`.
- **A4 — No vulnerability scanning in CI** `[Sonnet]`. Add `govulncheck` (build from source with the module toolchain, like the golangci-lint step) and/or a CodeQL workflow.

## B. Data model & API correctness
- **B1 — No `created_at`/`updated_at`** `[Opus]`. Ordering is by UUID `id` (arbitrary). Add timestamp columns (migration `0002`) and order by `created_at DESC` for "newest first"; removes the post-create-reset workaround in the UI.
- **B2 — `PUT` is full-replace, no `PATCH`** `[Opus]`. Omitted fields get zeroed — safe for the UI (sends the whole object) but a footgun for other clients. Add partial-update semantics or document the contract prominently.
- **B3 — Thin validation** `[Sonnet]`. `service.go` only checks `Title == ""`. Trim the title (whitespace-only currently passes), require `project_id`, and bound `priority` (0–3) / non-negative `story_points`/`estimate_hours`.
- **B4 — Dead entity fields** `[Opus]`. `entities.go` stores `parent_id, column_id, assignee_id, reporter_id, sprint_id, estimate_hours, epic_color` — none are exposed by any endpoint or the UI. Decide: build features around them (hierarchy, boards, assignees) or trim to what's used.
- **B5 — SQLite write concurrency** `[Sonnet]`. WAL + `busy_timeout` help, but the default pool can still hit `SQLITE_BUSY` under concurrent writers. Consider `SetMaxOpenConns(1)` or a write mutex, and document the single-instance assumption.

## C. Frontend
- **C1 — `description_markdown` rendered as plain text** `[Sonnet]`. `WorkItemCard.tsx` shows it raw. Either render markdown (e.g. `solid-markdown`) or rename the field to `description`.
- **C2 — No routing / multi-project navigation** `[Opus]`. Project is a free-text input; add a router + project list/switcher.
- **C3 — Theme code untested** `[Sonnet]`. `theme.ts`, `ThemeToggle.tsx`, and `App.tsx` have no tests. Add Vitest coverage (mode resolution, palette persistence, toggle interaction).
- **C4 — Thin pending affordance** `[Sonnet]`. `busy()` only disables "Add"; edit/delete give no spinner/disabled feedback.
- **C5 — Timezone edge** `[Sonnet]`. Dates are UTC-midnight; users in negative offsets can see an off-by-one due date. Pick local vs UTC and make it consistent.

## D. Testing
- **D1 — Presenter + edge cases untested (backend)** `[Sonnet]`. Add tests for `api/presenter` and body-parse failure paths.
- **D2 — No true end-to-end test** `[Opus]`. Everything is unit/component. Add one integration test that boots the real embedded binary and drives the HTTP API (or Playwright against the served SPA — Chromium is available in the environment).

## E. Ops / CI / DX
- **E1 — golangci-lint rebuilt from source every CI run** (~1–2 min) `[Sonnet]`. `.github/workflows/ci.yml` — cache the built binary (or the module/build caches) keyed on the pinned version.
- **E2 — No Docker build in CI** `[Sonnet]`. Add a job that `docker build`s (and optionally publishes) the image so the multi-stage Dockerfile is exercised on every PR.
- **E3 — No release/versioning** `[Sonnet]`. No tags/changelog; add lightweight release automation if this becomes a template others consume.
- **E4 — TypeScript held at 5.x** `[Sonnet]`. `typescript-eslint` does not support TypeScript ≥ 7 yet (the native compiler; typescript-eslint#10940). A Dependabot `ignore` keeps TS on 5.x for now — drop it and bump once the lint toolchain supports the new compiler.

## F. Docs & hygiene
- **F1 — No `LICENSE` file** though `package.json` declares MIT `[Sonnet]`. Add `LICENSE`. Consider also `SECURITY.md`, `CONTRIBUTING.md`, issue/PR templates, and `CODEOWNERS`.
- **F2 — No API reference** `[Opus]`. Add an OpenAPI spec (or a curl-based API doc) for the work-items endpoints + the `{status, data, error, meta}` envelope.

---

## Suggested order
1. **Quick wins**: F1 (`LICENSE`), B3 (validation), A3 (rate limit).
2. **UX unlock**: B1 (`created_at` + "newest first" ordering).
3. **Production blocker**: A1 (auth + tenancy).
4. **Confidence**: D2 (end-to-end test), A4 (govulncheck/CodeQL), E2 (Docker build in CI).
