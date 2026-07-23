package work_item

import (
	"errors"
	"testing"
)

func newTestService() Service {
	return NewService(NewInMemoryRepository())
}

func TestCreateWorkItem_RequiresTitle(t *testing.T) {
	svc := newTestService()
	if _, err := svc.CreateWorkItem(&WorkItem{ProjectID: "p1"}); !errors.Is(err, ErrTitleRequired) {
		t.Fatalf("expected ErrTitleRequired, got %v", err)
	}
}

func TestCreateWorkItem_AssignsIDAndPersists(t *testing.T) {
	svc := newTestService()

	created, err := svc.CreateWorkItem(&WorkItem{Title: "Ship it", ProjectID: "p1", ID: "client-supplied"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" || created.ID == "client-supplied" {
		t.Fatalf("expected server-assigned ID, got %q", created.ID)
	}

	got, err := svc.GetWorkItemByID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Ship it" {
		t.Fatalf("expected title %q, got %q", "Ship it", got.Title)
	}
}

func TestGetWorkItemByID_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetWorkItemByID("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateWorkItem_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.UpdateWorkItem(&WorkItem{ID: "missing", Title: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteWorkItem_NotFound(t *testing.T) {
	svc := newTestService()
	if err := svc.DeleteWorkItem("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListWorkItemsByProjectID_Filters(t *testing.T) {
	svc := newTestService()
	if _, err := svc.CreateWorkItem(&WorkItem{Title: "a", ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateWorkItem(&WorkItem{Title: "b", ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateWorkItem(&WorkItem{Title: "c", ProjectID: "p2"}); err != nil {
		t.Fatal(err)
	}

	items, err := svc.ListWorkItemsByProjectID("p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items for p1, got %d", len(items))
	}
}
