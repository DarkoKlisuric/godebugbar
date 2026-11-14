package godebugbar_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestProfilerMiddleware(t *testing.T) {
	t.Parallel()

	p := godebugbar.New()
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><p>Page</p></body></html>`))
	})

	handler := p.Middleware(app)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Debugbar-ID") == "" {
		t.Fatalf("expected X-Debugbar-ID header on response")
	}
	if !strings.Contains(rec.Body.String(), "/_debugbar/assets/toolbar.js") {
		t.Fatalf("expected toolbar script snippet in HTML response")
	}
}

func TestProfilerDisabledPassthrough(t *testing.T) {
	t.Parallel()

	p := godebugbar.New(godebugbar.WithEnabled(false))
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>hello</body></html>"))
	})

	handler := p.Middleware(app)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Debugbar-ID") != "" {
		t.Fatalf("expected disabled profiler not to set header")
	}
	if strings.Contains(rec.Body.String(), "toolbar.js") {
		t.Fatalf("expected disabled profiler not to inject script")
	}
}
