package engine_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestEngineMiddlewareAndEndpoints(t *testing.T) {
	t.Parallel()

	eng := engine.New()
	eng.SetStorageSize(10)
	eng.EnableHTMLInjection()
	eng.AddCollectors(
		collector.NewRequestCollector(),
		collector.NewQueryCollector(),
		collector.NewRuntimeCollector(),
		collector.NewRequestCollector(),
	)
	eng.AddExcludedPrefixes("/healthz")
	eng.AddExcludedPaths("/exact-skip")
	eng.SetCORSOrigins("*")
	eng.Finalize()

	owner := &struct{}{}
	engine.Register(owner, eng)
	if engine.Lookup(owner) != eng {
		t.Fatalf("expected Lookup to return registered engine")
	}
	if engine.Lookup(&struct{}{}) != nil {
		t.Fatalf("expected Lookup for unknown key to return nil")
	}

	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hello":"world"}`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><p>Page</p></body></html>`))
	})

	handler := eng.Middleware(app)

	htmlReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/home", nil)
	htmlRec := httptest.NewRecorder()
	handler.ServeHTTP(htmlRec, htmlReq)

	profileID := htmlRec.Header().Get(engine.ProfileHeader)
	if profileID == "" {
		t.Fatalf("expected %s header on HTML response", engine.ProfileHeader)
	}
	if htmlRec.Header().Get("Access-Control-Expose-Headers") != engine.ProfileHeader {
		t.Fatalf("expected Access-Control-Expose-Headers to contain %s", engine.ProfileHeader)
	}
	if !strings.Contains(htmlRec.Body.String(), "/_debugbar/assets/toolbar.js") {
		t.Fatalf("expected toolbar script snippet in HTML response, got %s", htmlRec.Body.String())
	}

	dashReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar", nil)
	dashRec := httptest.NewRecorder()
	handler.ServeHTTP(dashRec, dashReq)
	if dashRec.Code != http.StatusOK || !strings.Contains(dashRec.Body.String(), "GoDebugBar Profiler") {
		t.Fatalf("expected dashboard HTML at /_debugbar, got %d", dashRec.Code)
	}
	if dashRec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected CORS header on /_debugbar endpoint")
	}

	excludedReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	excludedRec := httptest.NewRecorder()
	handler.ServeHTTP(excludedRec, excludedReq)
	if excludedRec.Header().Get(engine.ProfileHeader) != "" {
		t.Fatalf("expected excluded prefix not to be profiled")
	}

	exactReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/exact-skip", nil)
	exactRec := httptest.NewRecorder()
	handler.ServeHTTP(exactRec, exactReq)
	if exactRec.Header().Get(engine.ProfileHeader) != "" {
		t.Fatalf("expected excluded path not to be profiled")
	}

	optReq := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/home", nil)
	optRec := httptest.NewRecorder()
	handler.ServeHTTP(optRec, optReq)
	if optRec.Header().Get(engine.ProfileHeader) != "" {
		t.Fatalf("expected OPTIONS request not to be profiled by default")
	}

	assetReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/assets/toolbar.js", nil)
	assetRec := httptest.NewRecorder()
	handler.ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK || !strings.Contains(assetRec.Body.String(), "GoDebugBar") {
		t.Fatalf("expected toolbar.js asset 200, got %d", assetRec.Code)
	}

	latestReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil)
	latestRec := httptest.NewRecorder()
	handler.ServeHTTP(latestRec, latestReq)
	if latestRec.Code != http.StatusOK {
		t.Fatalf("expected latest profile 200, got %d", latestRec.Code)
	}

	byIDReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID, nil)
	byIDRec := httptest.NewRecorder()
	handler.ServeHTTP(byIDRec, byIDReq)
	if byIDRec.Code != http.StatusOK {
		t.Fatalf("expected profile by ID 200, got %d", byIDRec.Code)
	}

	var prof engine.Profile
	if err := json.Unmarshal(byIDRec.Body.Bytes(), &prof); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if prof.ID != profileID || len(prof.Sections) != 3 {
		t.Fatalf("expected 3 deduplicated sections on profile, got %+v", prof)
	}

	missingReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/unknown", nil)
	missingRec := httptest.NewRecorder()
	handler.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown profile, got %d", missingRec.Code)
	}
}
