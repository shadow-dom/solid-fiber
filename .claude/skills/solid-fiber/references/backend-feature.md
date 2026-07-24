# Adding / extending a backend resource

The reference resource is `work_item` (`backend/pkg/work_item/`). To add a new
resource, replicate its layers. Below, `<resource>` is the snake_case package
name (e.g. `tag`), `<Resource>` the exported type (e.g. `Tag`). Read the real
`work_item` files alongside these templates — they are the source of truth.

The `//` notes in the templates below are guidance to *you*, not comments to ship.
The code you write should follow the project standard: self-documenting, with
comments only for non-obvious *why* (see SKILL.md → Core principles).

## Order of work — test-first

Define the contracts (types + interfaces) just enough to compile, then for each
layer **write the test, watch it fail, implement, refactor**:

1. Entity + `Repository`/`Service` interfaces (compile-only skeleton).
2. Service test → service + in-memory repository.
3. SQLite repository test → SQLite impl + migration.
4. Handler test (`app.Test`) → handlers + presenter + routes + wiring.

The templates are grouped by file below; reach for each when its test drives you there.

## 1. Entity — `pkg/<resource>/entities.go`

```go
package <resource>

// <Resource> is the core domain entity. Identity is a server-assigned UUID.
type <Resource> struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	// ... fields. Use `,omitempty` on optional fields; []string for lists.
}
```

## 2. Repository — `pkg/<resource>/repository.go`

The interface plus a goroutine-safe in-memory implementation (used by unit tests and as a drop-in). Keep method names identical to `work_item`.

```go
package <resource>

import (
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("<resource> not found")

type Repository interface {
	Create(item *<Resource>) (*<Resource>, error)
	GetByID(id string) (*<Resource>, error)
	Update(item *<Resource>) (*<Resource>, error)
	Delete(id string) error
	ListByProjectID(projectID string, limit, offset int) ([]*<Resource>, error)
	CountByProjectID(projectID string) (int, error)
}

type inMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*<Resource>
}

func NewInMemoryRepository() Repository {
	return &inMemoryRepository{items: make(map[string]*<Resource>)}
}
// Implement the methods exactly as work_item/repository.go does: clone on the
// way in and out, sort by ID for stable pagination, ErrNotFound on misses.
```

## 3. SQLite repository — `pkg/<resource>/sqlite_repository.go`

Uses only `database/sql` (the driver import lives in `pkg/storage`). Store slices as JSON text; booleans as INTEGER. Mirror `work_item/sqlite_repository.go`: a shared column list constant, `LIMIT ? OFFSET ?` with `ORDER BY id`, a `COUNT(*)` for `CountByProjectID`, and `NewSQLiteRepository` that runs migrations:

```go
func NewSQLiteRepository(db *sql.DB) (Repository, error) {
	if err := storage.Migrate(db, Migrations); err != nil {
		return nil, fmt.Errorf("migrate <resource> schema: %w", err)
	}
	return &sqliteRepository{db: db}, nil
}
```

## 4. Migration — `pkg/<resource>/migrations.go`

Append-only. Version numbers are per-resource (they live in their own `schema_migrations` table row set only if you share a DB — in this repo each resource owns its tables; keep versions monotonic within the slice).

```go
package <resource>

import "github.com/shadow-dom/solid-fiber/pkg/storage"

var Migrations = []storage.Migration{
	{
		Version: 1,
		Name:    "create_<resource>s",
		SQL: `
CREATE TABLE IF NOT EXISTS <resource>s (
	id         TEXT PRIMARY KEY,
	name       TEXT NOT NULL,
	project_id TEXT NOT NULL DEFAULT ''
	-- ... columns, NOT NULL DEFAULT for every field
);
CREATE INDEX IF NOT EXISTS idx_<resource>s_project_id ON <resource>s(project_id);
`,
	},
}
```

To evolve an existing resource, **add** a `{Version: 2, ...}` entry with `ALTER TABLE ...`; never edit version 1.

## 5. Service — `pkg/<resource>/service.go`

Use cases + validation + UUID assignment. Validation errors are sentinels the handler maps to 400.

```go
var (
	ErrNameRequired      = errors.New("name is required")
	ErrProjectIDRequired = errors.New("project_id is required")
)

func validate(w *<Resource>) error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" { return ErrNameRequired }
	if strings.TrimSpace(w.ProjectID) == "" { return ErrProjectIDRequired }
	return nil
}

type Service interface {
	Create<Resource>(*<Resource>) (*<Resource>, error)
	Get<Resource>ByID(id string) (*<Resource>, error)
	Update<Resource>(*<Resource>) (*<Resource>, error)
	Delete<Resource>(id string) error
	List<Resource>sByProjectID(projectID string, limit, offset int) ([]*<Resource>, int, error)
}

func (s *service) Create<Resource>(w *<Resource>) (*<Resource>, error) {
	if err := validate(w); err != nil { return nil, err }
	w.ID = uuid.NewString() // server-assigned; ignore client id
	return s.repo.Create(w)
}
// List returns (items, total, error): repo.ListByProjectID + repo.CountByProjectID.
```

## 6. Presenter — `api/presenter/<resource>.go`

```go
func <Resource>SuccessResponse(data *<resource>.<Resource>) fiber.Map {
	return fiber.Map{"status": true, "data": data, "error": nil}
}
func <Resource>sPaginatedResponse(data []*<resource>.<Resource>, total, limit, offset int) fiber.Map {
	return fiber.Map{"status": true, "data": data,
		"meta": fiber.Map{"total": total, "limit": limit, "offset": offset}, "error": nil}
}
func <Resource>ErrorResponse(err error) fiber.Map {
	return fiber.Map{"status": false, "data": nil, "error": err.Error()}
}
```

## 7. Handlers — `api/handlers/<resource>_handler.go`

Thin. Parse with `c.Bind().Body(&x)`, set the path id as source of truth on update, map errors with a `statusForError` (add this resource's validation sentinels), return `201` on create and `204` on delete, parse `limit`/`offset` (default 20, max 100) for list. Copy `work_item_handler.go` verbatim and rename.

## 8. Routes — `api/routes/<resource>.go`

```go
func <Resource>Router(router fiber.Router, service <resource>.Service) {
	g := router.Group("/<resource>s")
	g.Post("", handlers.Add<Resource>(service))
	g.Get("", handlers.List<Resource>s(service))
	g.Get("/:id", handlers.Get<Resource>(service))
	g.Put("/:id", handlers.Update<Resource>(service))
	g.Delete("/:id", handlers.Delete<Resource>(service))
}
```

## 9. Wire it up

- `main.go`: build the repo + service (`repo, _ := <resource>.NewSQLiteRepository(db); svc := <resource>.NewService(repo)`), pass into `api.New`.
- `api/server.go`: add a field to `Config`, then `routes.<Resource>Router(api, cfg.<Resource>s)` **before** the `/api` 404 catch-all `api.Use(...)`.

## The tests (write these first — they drive steps above)

Per the test-first order, each of these is written before the code it exercises,
mirroring `work_item`'s tests:

- **Service** (before steps 5/2): table-driven validation, create-assigns-UUID, not-found.
- **SQLite repository** (before step 3): CRUD + a persist-across-reopen check, using a temp DB — `storage.OpenSQLite(filepath.Join(t.TempDir(), "test.db"))`.
- **Handler** (before step 7): status codes, envelope shape, and pagination via `app.Test`.

## Then verify

Run the backend gates (see SKILL.md → Verification gates). A missing repository method on one of the two impls, or a route not registered, are the usual compile/runtime failures.
