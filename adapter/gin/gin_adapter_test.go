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
	Sections []godebugbar.Section `json:"sections"`
}

func TestGinAdapterCapturesRouteAndHTML(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	godebugbar.New(
		ginadapter.Attach(router),
		godebugbar.WithHTMLInjection(),
		godebugbar.WithCORS("*"),
	)

	router.GET("/users/:id", func(c *gin.Context) {
		_ = c.Error(errors.New("minor warning"))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h1>User</h1></body></html>"))
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/42", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	profileID := rec.Header().Get("X-Debugbar-ID")
	if profileID == "" {
		t.Fatalf("expected X-Debugbar-ID header")
	}
	if !strings.Contains(rec.Body.String(), "/_debugbar/assets/toolbar.js") {
		t.Fatalf("expected injected toolbar script, got %s", rec.Body.String())
	}

	dashReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar", nil)
	dashRec := httptest.NewRecorder()
	router.ServeHTTP(dashRec, dashReq)
	if dashRec.Code != http.StatusOK || !strings.Contains(dashRec.Body.String(), "GoDebugBar Profiler") {
		t.Fatalf("expected dashboard HTML at /_debugbar, got %d", dashRec.Code)
	}

	apiReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID, nil)
	apiRec := httptest.NewRecorder()
	router.ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusOK {
		t.Fatalf("expected debugbar API 200, got %d", apiRec.Code)
	}

	var prof profilePayload
	if err := json.Unmarshal(apiRec.Body.Bytes(), &prof); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if len(prof.Sections) != 1 || prof.Sections[0].Name != "request" {
		t.Fatalf("expected only request section from ginadapter.Attach, got %+v", prof.Sections)
	}

	raw := apiRec.Body.String()
	if !strings.Contains(raw, "/users/:id") {
		t.Fatalf("expected route pattern /users/:id in profile, got %s", raw)
	}
	if !strings.Contains(raw, "minor warning") {
		t.Fatalf("expected gin error in profile, got %s", raw)
	}
}
