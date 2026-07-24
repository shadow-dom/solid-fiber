package work_item_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/shadow-dom/solid-fiber/pkg/storage"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

func newSQLiteRepo(t *testing.T, path string) work_item.Repository {
	t.Helper()
	db, err := storage.OpenSQLite(path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo, err := work_item.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new sqlite repository: %v", err)
	}
	return repo
}

func TestSQLiteRepository_CRUDAndLabelsRoundTrip(t *testing.T) {
	repo := newSQLiteRepo(t, filepath.Join(t.TempDir(), "test.db"))

	in := &work_item.WorkItem{
		ID: "wi-1", Title: "Ship it", ProjectID: "p1",
		Priority: 3, EstimateHours: 2.5, IsMilestone: true,
		Labels: []string{"backend", "urgent"},
	}
	if _, err := repo.Create(in); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID("wi-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Ship it" || got.Priority != 3 || !got.IsMilestone {
		t.Fatalf("field round-trip mismatch: %+v", got)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "backend" || got.Labels[1] != "urgent" {
		t.Fatalf("labels round-trip mismatch: %v", got.Labels)
	}

	got.Title = "Shipped"
	got.Labels = nil
	if _, err := repo.Update(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	after, _ := repo.GetByID("wi-1")
	if after.Title != "Shipped" || len(after.Labels) != 0 {
		t.Fatalf("update not persisted: %+v", after)
	}

	if err := repo.Delete("wi-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID("wi-1"); !errors.Is(err, work_item.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSQLiteRepository_NotFound(t *testing.T) {
	repo := newSQLiteRepo(t, filepath.Join(t.TempDir(), "test.db"))
	if _, err := repo.GetByID("missing"); !errors.Is(err, work_item.ErrNotFound) {
		t.Fatalf("get: expected ErrNotFound, got %v", err)
	}
	if _, err := repo.Update(&work_item.WorkItem{ID: "missing", Title: "x"}); !errors.Is(err, work_item.ErrNotFound) {
		t.Fatalf("update: expected ErrNotFound, got %v", err)
	}
	if err := repo.Delete("missing"); !errors.Is(err, work_item.ErrNotFound) {
		t.Fatalf("delete: expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteRepository_ListByProjectIDAndCount(t *testing.T) {
	repo := newSQLiteRepo(t, filepath.Join(t.TempDir(), "test.db"))
	for _, wi := range []*work_item.WorkItem{
		{ID: "a", Title: "a", ProjectID: "p1"},
		{ID: "b", Title: "b", ProjectID: "p1"},
		{ID: "c", Title: "c", ProjectID: "p1"},
		{ID: "z", Title: "z", ProjectID: "p2"},
	} {
		if _, err := repo.Create(wi); err != nil {
			t.Fatal(err)
		}
	}

	count, err := repo.CountByProjectID("p1")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3 for p1, got %d", count)
	}

	// First page of 2, ordered by id -> a, b.
	page1, err := repo.ListByProjectID("p1", 2, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page1) != 2 || page1[0].ID != "a" || page1[1].ID != "b" {
		t.Fatalf("unexpected page 1: %+v", page1)
	}

	// Second page -> c.
	page2, err := repo.ListByProjectID("p1", 2, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page2) != 1 || page2[0].ID != "c" {
		t.Fatalf("unexpected page 2: %+v", page2)
	}
}

// TestSQLiteRepository_PersistsAcrossReopen is the core guarantee of this
// change: data written by one process is still there after the DB is closed
// and reopened from the same file.
func TestSQLiteRepository_PersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")

	repo := newSQLiteRepo(t, path)
	if _, err := repo.Create(&work_item.WorkItem{ID: "keep", Title: "durable", ProjectID: "p1"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Reopen the same file with a fresh connection/repository.
	reopened := newSQLiteRepo(t, path)
	got, err := reopened.GetByID("keep")
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if got.Title != "durable" {
		t.Fatalf("expected persisted item, got %+v", got)
	}
}
