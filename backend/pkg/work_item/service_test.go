package work_item

import (
	"errors"
	"testing"
)

func newTestService() Service {
	return NewService(NewInMemoryRepository())
}

func TestCreateWorkItem_Validation(t *testing.T) {
	cases := []struct {
		name string
		in   *WorkItem
		want error
	}{
		{"empty title", &WorkItem{ProjectID: "p1"}, ErrTitleRequired},
		{"whitespace title", &WorkItem{Title: "   ", ProjectID: "p1"}, ErrTitleRequired},
		{"missing project", &WorkItem{Title: "x"}, ErrProjectIDRequired},
		{"priority too high", &WorkItem{Title: "x", ProjectID: "p1", Priority: 4}, ErrInvalidPriority},
		{"priority negative", &WorkItem{Title: "x", ProjectID: "p1", Priority: -1}, ErrInvalidPriority},
		{"negative points", &WorkItem{Title: "x", ProjectID: "p1", StoryPoints: -2}, ErrNegativeValue},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService()
			if _, err := svc.CreateWorkItem(tc.in); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestCreateWorkItem_TrimsTitle(t *testing.T) {
	svc := newTestService()
	created, err := svc.CreateWorkItem(&WorkItem{Title: "  Trim me  ", ProjectID: "p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Title != "Trim me" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
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
	if _, err := svc.UpdateWorkItem(&WorkItem{ID: "missing", Title: "x", ProjectID: "p1"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteWorkItem_NotFound(t *testing.T) {
	svc := newTestService()
	if err := svc.DeleteWorkItem("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListWorkItemsByProjectID_FiltersAndPaginates(t *testing.T) {
	svc := newTestService()
	for _, title := range []string{"a", "b", "c"} {
		if _, err := svc.CreateWorkItem(&WorkItem{Title: title, ProjectID: "p1"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CreateWorkItem(&WorkItem{Title: "other", ProjectID: "p2"}); err != nil {
		t.Fatal(err)
	}

	// Full page: 3 items in p1, total 3.
	items, total, err := svc.ListWorkItemsByProjectID("p1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 || total != 3 {
		t.Fatalf("expected 3 items and total 3, got %d items total %d", len(items), total)
	}

	// First page of 2.
	page1, total, err := svc.ListWorkItemsByProjectID("p1", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 || total != 3 {
		t.Fatalf("expected 2 items and total 3, got %d items total %d", len(page1), total)
	}

	// Second page has the remaining 1.
	page2, _, err := svc.ListWorkItemsByProjectID("p1", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 item on page 2, got %d", len(page2))
	}
	// Pages must be disjoint (ordered by id).
	if page1[0].ID == page2[0].ID {
		t.Fatalf("pages overlap: %s", page2[0].ID)
	}
}
