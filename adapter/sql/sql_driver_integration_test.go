//go:build integration

package sqladapter_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DarkoKlisuric/godebugbar"
	sqladapter "github.com/DarkoKlisuric/godebugbar/adapter/sql"
)

func TestSQLAdapterPostgresIntegration(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	db, err := sqladapter.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	bar := godebugbar.New(godebugbar.WithEnabled(true), sqladapter.Attach())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var n int
		if err := db.QueryRowContext(r.Context(), "SELECT $1::int * 2", 21).Scan(&n); err != nil || n != 42 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/answer", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("query failed: %d", rec.Code)
	}
	apiRec := httptest.NewRecorder()
	handler.ServeHTTP(apiRec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))
	if body := apiRec.Body.String(); !strings.Contains(body, "SELECT $1::int * 2") || !strings.Contains(body, `"21"`) {
		t.Fatalf("expected the query in the profile, got %s", body)
	}
}
