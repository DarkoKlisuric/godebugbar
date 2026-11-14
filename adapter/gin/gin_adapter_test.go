package ginadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestGinAdapterAttach(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	godebugbar.New(
		Attach(engine),
	)

	engine.GET("/v1/plans/:id", func(c *gin.Context) {
		recorder := godebugbar.RecorderFromContext(c)
		if recorder == nil {
			t.Error("expected RecorderFromContext(gin.Context) to work via ContextWithFallback")
		}

		_ = c.Error(errors.New("sample warning"))
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	engine.GET("/page", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h1>Page</h1></body></html>"))
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/plans/pro", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	profileID := rec.Header().Get(godebugbar.ProfileHeader)
	if profileID == "" {
		t.Fatal("expected X-Debugbar-ID header on Gin response")
	}

	profileReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID, nil)
	profileRec := httptest.NewRecorder()
	engine.ServeHTTP(profileRec, profileReq)

	if profileRec.Code != http.StatusOK {
		t.Fatalf("GET profile status = %d, want 200", profileRec.Code)
	}

	var got godebugbar.Profile
	if err := json.Unmarshal(profileRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal profile error = %v", err)
	}

	if got.Route != "/v1/plans/:id" {
		t.Errorf("got.Route = %q, want /v1/plans/:id", got.Route)
	}

	htmlReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/page", nil)
	htmlRec := httptest.NewRecorder()
	engine.ServeHTTP(htmlRec, htmlReq)

	if !strings.Contains(htmlRec.Body.String(), "/_debugbar/toolbar.js") {
		t.Errorf("HTML response %q missing injected toolbar script", htmlRec.Body.String())
	}
}
