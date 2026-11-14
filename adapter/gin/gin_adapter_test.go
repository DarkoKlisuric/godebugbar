package ginadapter_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
	ginadapter "github.com/DarkoKlisuric/godebugbar/adapter/gin"
)

type profilePayload struct {
	ID       string               `json:"id"`
	Status   int                  `json:"status"`
	Route    string               `json:"route"`
	Sections []godebugbar.Section `json:"sections"`
}

func serve(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func latestProfile(t *testing.T, router *gin.Engine) profilePayload {
	t.Helper()

	var prof profilePayload
	rec := serve(t, router, "/_debugbar/api/profiles/latest")
	if err := json.Unmarshal(rec.Body.Bytes(), &prof); err != nil {
		t.Fatalf("decode latest profile %q: %v", rec.Body.String(), err)
	}

	return prof
}

func newRouter(t *testing.T, options ...godebugbar.Option) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	godebugbar.New(append([]godebugbar.Option{godebugbar.WithEnabled(true), ginadapter.Attach(router)}, options...)...)

	return router
}

func TestGinAdapterCapturesRouteAndHTML(t *testing.T) {
	t.Parallel()

	router := newRouter(t, godebugbar.WithHTMLInjection())
	router.GET("/users/:id", func(c *gin.Context) {
		_ = c.Error(errors.New("minor warning"))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body>"+strings.Repeat("İ", 3)+"<h1>User</h1></body></html>"))
	})

	rec := serve(t, router, "/users/42")
	if rec.Header().Get("X-Debugbar-ID") == "" {
		t.Fatalf("expected X-Debugbar-ID header")
	}
	if !strings.Contains(rec.Body.String(), `<h1>User</h1><script src="/_debugbar/assets/toolbar.js"`) {
		t.Fatalf("expected the toolbar before </body>, got %s", rec.Body.String())
	}
	if dash := serve(t, router, "/_debugbar"); dash.Code != http.StatusOK || !strings.Contains(dash.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("expected dashboard HTML at /_debugbar, got %d", dash.Code)
	}

	prof := latestProfile(t, router)
	raw, _ := json.Marshal(prof)
	if prof.Route != "/users/:id" || prof.Sections[0].Name != "request" || !strings.Contains(string(raw), "minor warning") {
		t.Fatalf("expected route, request section and gin error in the profile, got %s", raw)
	}
}

func TestGinAdapterRejectsExistingRoutes(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/early", func(c *gin.Context) { c.Status(http.StatusOK) })

	defer func() {
		if recover() == nil {
			t.Fatalf("expected Attach to panic when routes already exist")
		}
	}()
	godebugbar.New(godebugbar.WithEnabled(true), ginadapter.Attach(router))
}

func TestGinAdapterDisabledLeavesRouterAlone(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/early", func(c *gin.Context) { c.Status(http.StatusOK) })
	godebugbar.New(ginadapter.Attach(router))

	if rec := serve(t, router, "/early"); rec.Header().Get("X-Debugbar-ID") != "" {
		t.Fatalf("expected a disabled profiler not to touch the router")
	}
	if router.ContextWithFallback {
		t.Fatalf("expected ContextWithFallback to stay unchanged when disabled")
	}
}

func TestGinAdapterSetsContextFallback(t *testing.T) {
	t.Parallel()

	router := newRouter(t)
	if !router.ContextWithFallback {
		t.Fatalf("expected Attach to set ContextWithFallback")
	}

	var recorded bool
	router.GET("/ctx", func(c *gin.Context) {
		recorded = godebugbar.FromContext(c) != nil
		c.Status(http.StatusOK)
	})
	serve(t, router, "/ctx")
	if !recorded {
		t.Fatalf("expected FromContext(*gin.Context) to find the request recorder")
	}
}

func TestGinAdapterRecordsPanics(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	godebugbar.New(godebugbar.WithEnabled(true), ginadapter.Attach(router))
	router.GET("/explode", func(*gin.Context) { panic("boom") })

	if rec := serve(t, router, "/explode"); rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected gin.Recovery to answer 500, got %d", rec.Code)
	}
	prof := latestProfile(t, router)
	raw, _ := json.Marshal(prof)
	if prof.Status != http.StatusInternalServerError || !strings.Contains(string(raw), "Panic: boom") {
		t.Fatalf("expected a 500 profile with the panic, got %s", raw)
	}
}

func TestGinAdapterStreamsFlushedHTML(t *testing.T) {
	t.Parallel()

	router := newRouter(t, godebugbar.WithHTMLInjection())
	router.GET("/stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		_, _ = c.Writer.WriteString("<body>chunk")
		c.Writer.Flush()
		_, _ = c.Writer.WriteString("</body>")
	})

	rec := serve(t, router, "/stream")
	if !rec.Flushed || rec.Body.String() != "<body>chunk</body>" {
		t.Fatalf("expected a flushed response without injection, got flushed=%v %q", rec.Flushed, rec.Body.String())
	}
}
