package gormadapter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
)

type postgresAuditItem struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"size:120;not null"`
	CreatedAt time.Time
}

func TestGormPluginPostgresIntegration(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	tableName := fmt.Sprintf("godebugbar_items_%d", time.Now().UnixNano())
	scopedDB := db.Table(tableName)

	if err := scopedDB.AutoMigrate(&postgresAuditItem{}); err != nil {
		t.Fatalf("failed to migrate test table: %v", err)
	}
	t.Cleanup(func() {
		_ = scopedDB.Migrator().DropTable(&postgresAuditItem{})
	})

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		gormadapter.Attach(db),
	)
	router := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqDB := db.WithContext(r.Context()).Table(tableName)
		item := postgresAuditItem{Title: "postgres-check"}
		if err := reqDB.Create(&item).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var out []postgresAuditItem
		if err := reqDB.Where("title = ?", "postgres-check").Find(&out).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/audit", nil)
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
	if !strings.Contains(body, "postgres-check") || !strings.Contains(body, "INSERT INTO") || !strings.Contains(body, `"explainable":true`) {
		t.Fatalf("expected explainable Postgres SQL queries in profile JSON, got %s", body)
	}

	cases := []struct {
		name       string
		index      string
		wantStatus int
		wantBody   string
	}{
		{name: "insert is refused", index: "1", wantStatus: http.StatusBadRequest, wantBody: "only SELECT"},
		{name: "select is explained by postgres", index: "2", wantStatus: http.StatusOK, wantBody: "Scan"},
	}
	for _, tc := range cases {
		explainReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+profileID+"/queries/"+tc.index+"/explain", nil)
		explainRec := httptest.NewRecorder()
		router.ServeHTTP(explainRec, explainReq)
		if explainRec.Code != tc.wantStatus || !strings.Contains(explainRec.Body.String(), tc.wantBody) {
			t.Errorf("%s: explain = %d %s", tc.name, explainRec.Code, explainRec.Body.String())
		}
	}
}
