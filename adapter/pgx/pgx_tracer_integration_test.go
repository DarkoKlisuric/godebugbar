//go:build integration

package pgxadapter_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DarkoKlisuric/godebugbar"
	pgxadapter "github.com/DarkoKlisuric/godebugbar/adapter/pgx"
)

func TestPgxTracerPostgresIntegration(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	config.ConnConfig.Tracer = pgxadapter.NewTracer()
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	bar := godebugbar.New(godebugbar.WithEnabled(true), pgxadapter.Attach())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var n int
		if err := pool.QueryRow(r.Context(), "SELECT $1::int + 1", 41).Scan(&n); err != nil || n != 42 {
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
	if body := apiRec.Body.String(); !strings.Contains(body, "SELECT $1::int + 1") || !strings.Contains(body, `"41"`) {
		t.Fatalf("expected the traced query in the profile, got %s", body)
	}
}
