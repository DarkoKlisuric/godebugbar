package godebugbar_test

import (
	"encoding/json"
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
		godebugbar.WithRequestCollector(),
		godebugbar.WithRuntimeCollector(),
		godebugbar.WithDashboard(true),
		godebugbar.WithCORS("http://localhost:5173"),
		godebugbar.WithStorageSize(25),
		godebugbar.WithPrefix("custom_debug/"),
		godebugbar.WithExcludedPrefixes("/health"),
		godebugbar.WithExcludedPaths("/ready"),
		godebugbar.WithExcludedMethods(http.MethodOptions, http.MethodHead),
		godebugbar.WithCollectors(&stubCollector{}),
	)

	handler := p.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/page", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "toolbar.js") {
		t.Fatalf("expected HTML injection to be off unless WithHTMLInjection() is passed")
	}

	profileID := rec.Header().Get("X-Debugbar-ID")
	if profileID == "" {
		t.Fatalf("expected X-Debugbar-ID header")
	}

	apiReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/custom_debug/api/profiles/"+profileID, nil)
	apiReq.Header.Set("Origin", "http://localhost:5173")
	apiRec := httptest.NewRecorder()
	handler.ServeHTTP(apiRec, apiReq)
	if apiRec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("expected CORS origin header on debugbar API")
	}

	var decoded struct {
		Sections []godebugbar.Section `json:"sections"`
	}
	if err := json.Unmarshal(apiRec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if len(decoded.Sections) != 3 {
		t.Fatalf("expected 3 sections (request + runtime + custom stub), got %d", len(decoded.Sections))
	}

	healthReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Header().Get("X-Debugbar-ID") != "" {
		t.Fatalf("expected excluded prefix not to be instrumented")
	}

	readyReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
	readyRec := httptest.NewRecorder()
	handler.ServeHTTP(readyRec, readyReq)
	if readyRec.Header().Get("X-Debugbar-ID") != "" {
		t.Fatalf("expected excluded path not to be instrumented")
	}

	headReq := httptest.NewRequestWithContext(t.Context(), http.MethodHead, "/page", nil)
	headRec := httptest.NewRecorder()
	handler.ServeHTTP(headRec, headReq)
	if headRec.Header().Get("X-Debugbar-ID") != "" {
		t.Fatalf("expected excluded method not to be instrumented")
	}
}
