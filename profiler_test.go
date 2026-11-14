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

	cases := []struct {
		name       string
		options    []godebugbar.Option
		wantHeader bool
		wantScript bool
	}{
		{name: "disabled by default", wantHeader: false, wantScript: false},
		{name: "enabled", options: []godebugbar.Option{godebugbar.WithEnabled(true)}, wantHeader: true, wantScript: false},
		{name: "enabled with injection", options: []godebugbar.Option{godebugbar.WithEnabled(true), godebugbar.WithHTMLInjection()}, wantHeader: true, wantScript: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := godebugbar.New(tc.options...)
			handler := p.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(`<html><body><p>Page</p></body></html>`))
			}))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/home", nil))

			if got := rec.Header().Get("X-Debugbar-ID") != ""; got != tc.wantHeader {
				t.Fatalf("header present = %v, want %v", got, tc.wantHeader)
			}
			if got := strings.Contains(rec.Body.String(), "/_debugbar/assets/toolbar.js"); got != tc.wantScript {
				t.Fatalf("script present = %v, want %v", got, tc.wantScript)
			}
			if p.Enabled() != tc.wantHeader {
				t.Fatalf("Enabled() = %v", p.Enabled())
			}
		})
	}
}

func TestProfilerHandler(t *testing.T) {
	t.Parallel()

	p := godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithServeEndpoints(false))
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("expected an empty list from Handler, got %d %q", rec.Code, rec.Body.String())
	}
}
