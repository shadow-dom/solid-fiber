package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthURL(t *testing.T) {
	cases := map[string]string{
		":3000":            "http://127.0.0.1:3000/api/health",
		"0.0.0.0:8080":     "http://127.0.0.1:8080/api/health",
		"example.com:3000": "http://example.com:3000/api/health",
	}
	for in, want := range cases {
		if got := healthURL(in); got != want {
			t.Errorf("healthURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRunHealthcheck(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	if err := runHealthcheck(ok.URL); err != nil {
		t.Fatalf("expected healthy, got %v", err)
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer bad.Close()
	if err := runHealthcheck(bad.URL); err == nil {
		t.Fatal("expected error for 503 response, got nil")
	}

	// Nothing listening -> connection error.
	if err := runHealthcheck("http://127.0.0.1:0/api/health"); err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}
