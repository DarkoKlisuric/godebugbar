package gormadapter_test

import (
	"encoding/json"
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

func latestBody(t *testing.T, router *gin.Engine) string {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec.Body.String()
}

func TestGormPluginRecordsQueries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		options     []godebugbar.Option
		useGinCtx   bool
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "request context with values",
			wantPresent: []string{"zebra", "yak", "gorm_plugin_test.go", "N+1 (3×)"},
		},
		{
			name:        "gin context works as the query context",
			useGinCtx:   true,
			wantPresent: []string{"zebra", "yak"},
		},
		{
			name:        "hidden values keep only placeholders",
			options:     []godebugbar.Option{godebugbar.WithHiddenSQLValues()},
			wantPresent: []string{`name = ?`},
			wantAbsent:  []string{"zebra", "yak"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			router := gin.New()
			godebugbar.New(append([]godebugbar.Option{
				godebugbar.WithEnabled(true),
				ginadapter.Attach(router),
				gormadapter.Attach(db),
			}, tc.options...)...)

			router.GET("/items", func(c *gin.Context) {
				var tx *gorm.DB
				if tc.useGinCtx {
					tx = db.WithContext(c)
				} else {
					tx = db.WithContext(c.Request.Context())
				}
				for range 3 {
					var items []sampleModel
					_ = tx.Where("name = ?", "zebra").Find(&items).Error
				}
				_ = tx.Create(&sampleModel{Name: "yak"}).Error
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil)
			router.ServeHTTP(httptest.NewRecorder(), req)

			body := latestBody(t, router)
			var decoded map[string]any
			if err := json.Unmarshal([]byte(body), &decoded); err != nil {
				t.Fatalf("decode: %v", err)
			}
			for _, want := range tc.wantPresent {
				if !strings.Contains(body, want) {
					t.Errorf("expected %q in profile %s", want, body)
				}
			}
			for _, unwanted := range tc.wantAbsent {
				if strings.Contains(body, unwanted) {
					t.Errorf("expected %q not in profile", unwanted)
				}
			}
		})
	}
}

func TestGormPluginStandalone(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Use(gormadapter.Plugin()); err != nil {
		t.Fatalf("Use(Plugin()): %v", err)
	}
	bar := godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithQueryCollector())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var items []sampleModel
		_ = db.WithContext(r.Context()).Where("name = ?", "standalone").Find(&items).Error
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))
	if !strings.Contains(rec.Body.String(), "standalone") {
		t.Fatalf("expected the standalone plugin to record the query, got %s", rec.Body.String())
	}
}
