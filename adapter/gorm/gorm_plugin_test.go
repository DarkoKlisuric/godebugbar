package gormadapter_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/DarkoKlisuric/godebugbar"
	ginadapter "github.com/DarkoKlisuric/godebugbar/adapter/gin"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
)

type sampleModel struct {
	ID   uint
	Name string
}

func TestGormPluginRecordsDryRunQueries(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected gorm open error: %v", err)
	}

	router := gin.New()
	godebugbar.New(
		ginadapter.Attach(router),
		gormadapter.Attach(db),
	)

	router.GET("/items", func(c *gin.Context) {
		var items []sampleModel
		_ = db.WithContext(c.Request.Context()).Where("name = ?", "widget").Find(&items).Error
		_ = db.WithContext(c.Request.Context()).Create(&sampleModel{Name: "gadget"}).Error
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	profileID := rec.Header().Get("X-Debugbar-ID")
	if profileID == "" {
		t.Fatalf("expected X-Debugbar-ID header")
	}

	apiReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID, nil)
	apiRec := httptest.NewRecorder()
	router.ServeHTTP(apiRec, apiReq)

	body := apiRec.Body.String()
	if !strings.Contains(body, "widget") || !strings.Contains(body, "gadget") {
		t.Fatalf("expected recorded dry-run queries in profile response, got %s", body)
	}
}
