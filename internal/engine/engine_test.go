package engine_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

const appOrigin = "http://localhost:5173"

func newEngine(t *testing.T, configure func(*engine.Engine)) *engine.Engine {
	t.Helper()

	eng := engine.New()
	eng.SetEnabled(true)
	if configure != nil {
		configure(eng)
	}
	eng.Finalize(collector.NewRequestCollector())

	return eng
}

func serve(t *testing.T, h http.Handler, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func htmlApp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hello":"world"}`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><p>Page</p></body></html>`))
	})
}

func TestEngineDisabledByDefault(t *testing.T) {
	t.Parallel()

	eng := engine.New()
	eng.Finalize()
	app := htmlApp()

	if got := eng.Middleware(app); got == nil || eng.Enabled() {
		t.Fatalf("expected a new engine to be disabled")
	}
	rec := serve(t, eng.Middleware(app), http.MethodGet, "/_debugbar", nil)
	if rec.Header().Get(engine.ProfileHeader) != "" || strings.Contains(rec.Body.String(), "GoDebugBar") {
		t.Fatalf("expected a disabled engine to pass requests through untouched")
	}
}

func TestEngineEndpoints(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, func(e *engine.Engine) {
		e.EnableHTMLInjection()
		e.AddExcludedPrefixes("/healthz")
		e.AddExcludedPaths("/exact-skip")
	})
	handler := eng.Middleware(htmlApp())

	page := serve(t, handler, http.MethodGet, "/home", nil)
	profileID := page.Header().Get(engine.ProfileHeader)
	if profileID == "" || !strings.Contains(page.Body.String(), "/_debugbar/assets/toolbar.js") {
		t.Fatalf("expected profile header and injected toolbar, got %q", page.Body.String())
	}

	cases := []struct {
		name       string
		target     string
		wantStatus int
		wantBody   string
	}{
		{name: "dashboard", target: "/_debugbar", wantStatus: http.StatusOK, wantBody: "<!DOCTYPE html>"},
		{name: "dashboard with slash", target: "/_debugbar/", wantStatus: http.StatusOK, wantBody: "<!DOCTYPE html>"},
		{name: "toolbar asset", target: "/_debugbar/assets/toolbar.js", wantStatus: http.StatusOK, wantBody: "godebugbar-host"},
		{name: "render asset", target: "/_debugbar/assets/render.js", wantStatus: http.StatusOK, wantBody: "GoDebugBarRender"},
		{name: "list returns summaries", target: "/_debugbar/api/profiles", wantStatus: http.StatusOK, wantBody: `"badges"`},
		{name: "latest", target: "/_debugbar/api/profiles/latest", wantStatus: http.StatusOK, wantBody: profileID},
		{name: "by id", target: "/_debugbar/api/profiles/" + profileID, wantStatus: http.StatusOK, wantBody: `"sections"`},
		{name: "unknown id", target: "/_debugbar/api/profiles/unknown", wantStatus: http.StatusNotFound, wantBody: "profile not found"},
		{name: "explain without explainer", target: "/_debugbar/api/profiles/" + profileID + "/queries/1/explain", wantStatus: http.StatusNotFound, wantBody: "no explainer"},
		{name: "unknown endpoint", target: "/_debugbar/nope", wantStatus: http.StatusNotFound, wantBody: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, handler, http.MethodGet, tc.target, nil)
			if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("GET %s = %d %q, want %d containing %q", tc.target, rec.Code, rec.Body.String(), tc.wantStatus, tc.wantBody)
			}
		})
	}

	skipped := []struct {
		name   string
		method string
		target string
	}{
		{name: "excluded prefix", method: http.MethodGet, target: "/healthz/live"},
		{name: "excluded path", method: http.MethodGet, target: "/exact-skip"},
		{name: "options is always excluded", method: http.MethodOptions, target: "/home"},
	}
	for _, tc := range skipped {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if rec := serve(t, handler, tc.method, tc.target, nil); rec.Header().Get(engine.ProfileHeader) != "" {
				t.Fatalf("expected %s %s not to be profiled", tc.method, tc.target)
			}
		})
	}
}

func TestEngineListSince(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, nil)
	handler := eng.Middleware(htmlApp())
	first := serve(t, handler, http.MethodGet, "/one", nil).Header().Get(engine.ProfileHeader)
	serve(t, handler, http.MethodGet, "/two", nil)

	var list []engine.ProfileSummary
	rec := serve(t, handler, http.MethodGet, "/_debugbar/api/profiles?since="+first, nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 || list[0].Path != "/two" {
		t.Fatalf("expected only the newer profile, got %+v", list)
	}
}

func TestEngineAuthorizer(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, func(e *engine.Engine) {
		e.EnableHTMLInjection()
		e.SetAuthorizer(func(r *http.Request) bool { return r.Header.Get("X-Dev") == "yes" })
	})
	handler := eng.Middleware(htmlApp())

	denied := serve(t, handler, http.MethodGet, "/home", nil)
	if denied.Header().Get(engine.ProfileHeader) != "" || strings.Contains(denied.Body.String(), "toolbar.js") {
		t.Fatalf("expected no header and no toolbar for an unauthorized viewer")
	}
	if rec := serve(t, handler, http.MethodGet, "/_debugbar/api/profiles", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for an unauthorized API call, got %d", rec.Code)
	}

	allowed := map[string]string{"X-Dev": "yes"}
	if rec := serve(t, handler, http.MethodGet, "/home", allowed); rec.Header().Get(engine.ProfileHeader) == "" {
		t.Fatalf("expected the profile header for an authorized viewer")
	}
	if rec := serve(t, handler, http.MethodGet, "/_debugbar/api/profiles", allowed); rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for an authorized API call, got %d", rec.Code)
	}
}

func TestEngineCORS(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		origins    []string
		origin     string
		wantOrigin string
	}{
		{name: "listed origin is reflected", origins: []string{appOrigin}, origin: appOrigin, wantOrigin: appOrigin},
		{name: "unlisted origin is refused", origins: []string{appOrigin}, origin: "https://evil.example", wantOrigin: ""},
		{name: "wildcard is ignored", origins: []string{"*"}, origin: "https://evil.example", wantOrigin: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := newEngine(t, func(e *engine.Engine) { e.SetCORSOrigins(tc.origins...) })
			rec := serve(t, eng.Middleware(htmlApp()), http.MethodGet, "/_debugbar/api/profiles", map[string]string{"Origin": tc.origin})
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tc.wantOrigin)
			}
		})
	}
}

func TestEngineRedactsSecrets(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, nil)
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "SECRET-SESSION", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
		w.WriteHeader(http.StatusOK)
	})
	handler := eng.Middleware(app)
	serve(t, handler, http.MethodGet, "/me?token=SECRET-QUERY&page=2", map[string]string{
		"Authorization": "Bearer SECRET-JWT",
		"Cookie":        "sid=SECRET-COOKIE",
	})

	body := serve(t, handler, http.MethodGet, "/_debugbar/api/profiles/latest", nil).Body.String()
	for _, secret := range []string{"SECRET-JWT", "SECRET-COOKIE", "SECRET-SESSION", "SECRET-QUERY"} {
		if strings.Contains(body, secret) {
			t.Fatalf("expected %s to be redacted from the profile", secret)
		}
	}
	if !strings.Contains(body, "[redacted]") || !strings.Contains(body, "page") {
		t.Fatalf("expected redaction markers and untouched params, got %s", body)
	}
}

func TestEngineRecordsPanics(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, nil)
	handler := eng.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("expected the panic to be re-raised")
			}
		}()
		serve(t, handler, http.MethodGet, "/explode", nil)
	}()

	prof, ok := eng.Store().Latest()
	if !ok || prof.Status != http.StatusInternalServerError || prof.Path != "/explode" {
		t.Fatalf("expected a 500 profile for the panicking request, got %+v", prof)
	}
	raw, _ := json.Marshal(prof)
	if !strings.Contains(string(raw), "Panic: boom") {
		t.Fatalf("expected the panic in the request section, got %s", raw)
	}
}

func TestEngineCapturesRequestBody(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, func(e *engine.Engine) { e.SetRequestBodyLimit(1024) })
	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.c","password":"hunter2"}`))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	prof, _ := eng.Store().Latest()
	raw, _ := json.Marshal(prof)
	if strings.Contains(string(raw), "hunter2") || !strings.Contains(string(raw), "a@b.c") {
		t.Fatalf("expected the body with the password redacted, got %s", raw)
	}
}

func TestEngineExplain(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, func(e *engine.Engine) {
		e.SetExplainer(func(_ context.Context, sql string) ([]string, error) {
			if strings.Contains(sql, "broken") {
				return nil, errors.New("syntax error")
			}
			return []string{"Seq Scan on users"}, nil
		})
	})
	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := contextstate.RecorderFromContext(r.Context())
		rec.RecordQuery(contextstate.QueryEntry{SQL: "SELECT * FROM users"})
		rec.RecordQuery(contextstate.QueryEntry{SQL: "DELETE FROM users"})
		rec.RecordQuery(contextstate.QueryEntry{SQL: "SELECT broken"})
		w.WriteHeader(http.StatusOK)
	}))
	id := serve(t, handler, http.MethodGet, "/users", nil).Header().Get(engine.ProfileHeader)

	cases := []struct {
		name       string
		index      string
		wantStatus int
		wantBody   string
	}{
		{name: "select is explained", index: "1", wantStatus: http.StatusOK, wantBody: "Seq Scan"},
		{name: "delete is refused", index: "2", wantStatus: http.StatusBadRequest, wantBody: "only SELECT"},
		{name: "explainer error", index: "3", wantStatus: http.StatusUnprocessableEntity, wantBody: "syntax error"},
		{name: "index out of range", index: "9", wantStatus: http.StatusNotFound, wantBody: "query not found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, handler, http.MethodGet, "/_debugbar/api/profiles/"+id+"/queries/"+tc.index+"/explain", nil)
			if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("explain %s = %d %q", tc.index, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestEngineSeparateHandler(t *testing.T) {
	t.Parallel()

	eng := newEngine(t, func(e *engine.Engine) {
		e.SetServeEndpoints(false)
		e.SetPublicURL("http://localhost:9999/_debugbar/")
		e.EnableHTMLInjection()
	})
	app := eng.Middleware(htmlApp())

	page := serve(t, app, http.MethodGet, "/home", nil)
	if !strings.Contains(page.Body.String(), `src="http://localhost:9999/_debugbar/assets/toolbar.js"`) {
		t.Fatalf("expected the snippet to use the public URL, got %q", page.Body.String())
	}
	if rec := serve(t, app, http.MethodGet, "/_debugbar/api/profiles", nil); strings.Contains(rec.Body.String(), "badges") {
		t.Fatalf("expected the app middleware not to serve the API")
	}
	if rec := serve(t, eng.Handler(), http.MethodGet, "/_debugbar/api/profiles", nil); !strings.Contains(rec.Body.String(), "badges") {
		t.Fatalf("expected Handler to serve the API, got %q", rec.Body.String())
	}
}

type namedCollector struct{ name string }

func (c *namedCollector) Name() string { return c.name }

func (c *namedCollector) Collect(context.Context, *http.Request, *contract.ResponseInfo) *contract.Section {
	return nil
}

func TestEngineCollectorNames(t *testing.T) {
	t.Parallel()

	eng := engine.New()
	eng.AddCollectors(collector.NewRequestCollector(), collector.NewRequestCollector())
	if len(eng.Collectors()) != 1 {
		t.Fatalf("expected the same collector type to be registered once")
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("expected a panic for two different collectors named request")
		}
	}()
	eng.AddCollectors(&namedCollector{name: "request"})
}
