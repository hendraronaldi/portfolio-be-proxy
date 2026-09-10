package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestSessionMiddlewareGeneratesWhenMissing(t *testing.T) {
	var downstreamSeen string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamSeen = r.Header.Get("X-Session-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/agent/resume", nil)
	rec := httptest.NewRecorder()
	sessionMiddleware(next).ServeHTTP(rec, req)

	echoed := rec.Header().Get("X-Session-Id")
	if echoed == "" {
		t.Fatal("expected X-Session-Id to be echoed on response")
	}
	if !uuidV4Pattern.MatchString(echoed) {
		t.Fatalf("expected generated session id to be UUIDv4, got %q", echoed)
	}
	if downstreamSeen != echoed {
		t.Fatalf("expected downstream header %q to match echoed %q", downstreamSeen, echoed)
	}
}

func TestSessionMiddlewarePreservesWhenPresent(t *testing.T) {
	const existing = "123e4567-e89b-42d3-a456-426614174000"
	var downstreamSeen string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamSeen = r.Header.Get("X-Session-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/agent/resume", nil)
	req.Header.Set("X-Session-Id", existing)
	rec := httptest.NewRecorder()
	sessionMiddleware(next).ServeHTTP(rec, req)

	echoed := rec.Header().Get("X-Session-Id")
	if echoed != existing {
		t.Fatalf("expected echoed session id %q, got %q", existing, echoed)
	}
	if downstreamSeen != existing {
		t.Fatalf("expected downstream session id %q, got %q", existing, downstreamSeen)
	}
}

func TestSessionMiddlewareOptionsPassthrough(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodOptions, "/api/agent/resume", nil)
	rec := httptest.NewRecorder()
	sessionMiddleware(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called for OPTIONS")
	}
	if got := rec.Header().Get("X-Session-Id"); got == "" {
		t.Fatal("expected X-Session-Id header present on OPTIONS response")
	}
}
