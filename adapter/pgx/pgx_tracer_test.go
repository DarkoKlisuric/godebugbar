package pgxadapter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/DarkoKlisuric/godebugbar"
	pgxadapter "github.com/DarkoKlisuric/godebugbar/adapter/pgx"
)

func TestPgxTracerRecordsQueries(t *testing.T) {
	t.Parallel()

	tracer := pgxadapter.NewTracer()
	bar := godebugbar.New(godebugbar.WithEnabled(true), pgxadapter.Attach())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := tracer.TraceQueryStart(r.Context(), nil, pgx.TraceQueryStartData{SQL: "SELECT * FROM groups WHERE id = $1", Args: []any{42}})
		tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{CommandTag: pgconn.NewCommandTag("SELECT 1")})

		ctx = tracer.TraceQueryStart(r.Context(), nil, pgx.TraceQueryStartData{SQL: "DELETE FROM groups"})
		tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{Err: errors.New("permission denied")})
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/groups", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))

	for _, want := range []string{"SELECT * FROM groups WHERE id = $1", `"42"`, "permission denied", "pgx_tracer_test.go"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("expected %q in %s", want, rec.Body.String())
		}
	}
}

func TestPgxTracerIgnoresUnprofiledContext(t *testing.T) {
	t.Parallel()

	tracer := pgxadapter.NewTracer()
	ctx := tracer.TraceQueryStart(t.Context(), nil, pgx.TraceQueryStartData{SQL: "SELECT 1"})
	if ctx != t.Context() {
		t.Fatalf("expected the context to be returned unchanged outside a profiled request")
	}
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})
}
