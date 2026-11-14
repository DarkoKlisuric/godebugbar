package godebugbar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubCollector struct {
	name string
}

func (c *stubCollector) Name() string {
	return c.name
}

func (c *stubCollector) Collect(ctx context.Context, _ *http.Request, _ *ResponseInfo) *Section {
	recorder := RecorderFromContext(ctx)
	value := "none"
	if recorder != nil {
		if v, ok := recorder.Value("key"); ok {
			value, _ = v.(string)
		}
	}

	return NewSection(
		ToolbarItem{Label: "Stub", Value: value, Color: ColorGreen},
		NewPanel("Stub Panel"),
	)
}

func TestProfilerMiddlewareAndAPI(t *testing.T) {
	t.Parallel()

	profiler := NewProfiler(
		WithCollectors(&stubCollector{name: "stub"}),
		WithExcludedPrefixes("/healthz"),
	)

	appHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if recorder := RecorderFromContext(request.Context()); recorder != nil {
			recorder.SetValue("key", "recorded-value")
			recorder.SetRoute("/pages/{slug}")
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte("<html><body>ok</body></html>"))
	})

	wrapped := profiler.Middleware(appHandler)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/pages/home", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	profileID := rec.Header().Get(ProfileHeader)
	if profileID == "" {
		t.Fatal("expected X-Debugbar-ID header to be set")
	}

	if !strings.Contains(rec.Body.String(), `data-profile-id="`+profileID+`"`) {
		t.Errorf("HTML body %q missing injected toolbar script for profile %q", rec.Body.String(), profileID)
	}

	profileReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID, nil)
	profileRec := httptest.NewRecorder()
	wrapped.ServeHTTP(profileRec, profileReq)

	if profileRec.Code != http.StatusOK {
		t.Fatalf("GET profile status = %d, want 200", profileRec.Code)
	}

	var got Profile
	if err := json.Unmarshal(profileRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal profile error = %v", err)
	}

	if got.ID != profileID || got.Route != "/pages/{slug}" {
		t.Errorf("got profile = %+v, want ID=%q Route=/pages/{slug}", got, profileID)
	}

	if len(got.Sections) != 1 || got.Sections[0].Toolbar.Value != "recorded-value" {
		t.Errorf("got sections = %+v, want stub collector with recorded-value", got.Sections)
	}

	listReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles", nil)
	listRec := httptest.NewRecorder()
	wrapped.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("GET profiles list status = %d, want 200", listRec.Code)
	}

	excludedReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	excludedRec := httptest.NewRecorder()
	wrapped.ServeHTTP(excludedRec, excludedReq)

	if excludedRec.Header().Get(ProfileHeader) != "" {
		t.Error("expected excluded path /healthz not to have X-Debugbar-ID")
	}
}
