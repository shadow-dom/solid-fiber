package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v3"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/shadow-dom/solid-fiber/api/handlers"
	"github.com/shadow-dom/solid-fiber/pkg/work_item"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func testApp(pinger handlers.Pinger) *fiber.App {
	spa := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>spa</title>")},
	}
	return New(Config{
		WorkItems: work_item.NewService(work_item.NewInMemoryRepository()),
		Pinger:    pinger,
		SPA:       spa,
	})
}

func get(t *testing.T, app *fiber.App, target string) (*http.Response, map[string]any) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, target, nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	return resp, parsed
}

func TestHealth_OK(t *testing.T) {
	resp, body := get(t, testApp(fakePinger{}), "/api/health")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if body["status"] != true {
		t.Fatalf("expected status true, got %v", body)
	}
}

func TestHealth_Unavailable(t *testing.T) {
	resp, body := get(t, testApp(fakePinger{err: context.DeadlineExceeded}), "/api/health")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
	if body["status"] != false {
		t.Fatalf("expected status false, got %v", body)
	}
}

func TestUnknownAPIRoute_JSON404(t *testing.T) {
	resp, body := get(t, testApp(fakePinger{}), "/api/nope")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if body["status"] != false || body["error"] == nil || body["error"] == "" {
		t.Fatalf("expected JSON error envelope, got %v", body)
	}
}

func TestSPAFallback_ServesIndex(t *testing.T) {
	resp, _ := get(t, testApp(fakePinger{}), "/some/client/route")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct == "" || ct[:9] != "text/html" {
		t.Fatalf("expected html content type, got %q", ct)
	}
}

// TestPanicRecovered verifies the recover middleware + jsonErrorHandler turn a
// panic into a 500 JSON envelope rather than dropping the connection.
func TestPanicRecovered(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: jsonErrorHandler})
	app.Use(recoverer.New())
	app.Get("/boom", func(fiber.Ctx) error { panic("kaboom") })

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("expected JSON body, got %q", raw)
	}
	if body["status"] != false {
		t.Fatalf("expected status false, got %v", body)
	}
}
