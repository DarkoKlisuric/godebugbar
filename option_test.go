package godebugbar_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestOptionsConfiguration(t *testing.T) {
	t.Parallel()

	p := godebugbar.New(
		godebugbar.WithEnabled(true),
		godebugbar.WithDefaultCollectors(false),
		godebugbar.WithStorageSize(25),
		godebugbar.WithPrefix("custom_debug/"),
		godebugbar.WithExcludedPrefixes("/health"),
		godebugbar.WithCollectors(&stubCollector{}),
	)

	handler := p.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/page", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "/custom_debug/assets/toolbar.js") {
		t.Fatalf("expected custom prefix in injected snippet, got %s", rec.Body.String())
	}

	healthReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if strings.Contains(healthRec.Body.String(), "toolbar.js") {
		t.Fatalf("expected excluded prefix not to be instrumented")
	}
}
