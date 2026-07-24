package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/shadow-dom/solid-fiber/api/routes"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

func newTestApp() *fiber.App {
	app := fiber.New()
	svc := work_item.NewService(work_item.NewInMemoryRepository())
	routes.WorkItemRouter(app.Group("/api"), svc)
	return app
}

func do(t *testing.T, app *fiber.App, method, target, body string) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return resp.StatusCode, parsed
}

func TestAddWorkItem_EmptyTitle_BadRequest(t *testing.T) {
	app := newTestApp()
	status, _ := do(t, app, http.MethodPost, "/api/work-items", `{"project_id":"p1"}`)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestAddWorkItem_Created(t *testing.T) {
	app := newTestApp()
	status, body := do(t, app, http.MethodPost, "/api/work-items", `{"title":"Ship it","project_id":"p1"}`)
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["id"] == "" {
		t.Fatalf("expected data.id in response, got %v", body)
	}
}

func TestGetWorkItem_NotFound(t *testing.T) {
	app := newTestApp()
	status, _ := do(t, app, http.MethodGet, "/api/work-items/missing", "")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestWorkItem_RoundTrip(t *testing.T) {
	app := newTestApp()

	_, created := do(t, app, http.MethodPost, "/api/work-items", `{"title":"Ship it","project_id":"p1"}`)
	id, _ := created["data"].(map[string]any)["id"].(string)
	if id == "" {
		t.Fatalf("missing id in create response: %v", created)
	}

	status, got := do(t, app, http.MethodGet, "/api/work-items/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d", status)
	}
	if title := got["data"].(map[string]any)["title"]; title != "Ship it" {
		t.Fatalf("expected title round-trip, got %v", title)
	}

	status, _ = do(t, app, http.MethodDelete, "/api/work-items/"+id, "")
	if status != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", status)
	}

	status, _ = do(t, app, http.MethodGet, "/api/work-items/"+id, "")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", status)
	}
}

func TestUnknownAPIRoute_Returns404(t *testing.T) {
	app := newTestApp()
	// Register the same catch-all main.go uses.
	app.Group("/api").Use(func(c fiber.Ctx) error { return fiber.ErrNotFound })
	status, _ := do(t, app, http.MethodGet, "/api/does-not-exist", "")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}
