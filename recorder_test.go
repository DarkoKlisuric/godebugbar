package godebugbar

import (
	"context"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	t.Parallel()

	if got := RecorderFromContext(context.Background()); got != nil {
		t.Fatalf("RecorderFromContext(empty) = %v, want nil", got)
	}

	recorder := NewRecorder()
	ctx := WithRecorder(context.Background(), recorder)

	if got := RecorderFromContext(ctx); got != recorder {
		t.Fatalf("RecorderFromContext(ctx) = %v, want %v", got, recorder)
	}

	recorder.SetRoute("/v1/plans")
	if got := recorder.Route(); got != "/v1/plans" {
		t.Errorf("Route() = %q, want %q", got, "/v1/plans")
	}

	recorder.RecordQuery(QueryEntry{
		SQL:          "SELECT * FROM billing.plans",
		RowsAffected: 3,
		Duration:     2 * time.Millisecond,
	})
	queries := recorder.Queries()
	if len(queries) != 1 || queries[0].RowsAffected != 3 {
		t.Errorf("Queries() = %+v, want 1 query with 3 rows", queries)
	}

	endSpan := recorder.StartSpan("load plans", "app")
	endSpan()
	spans := recorder.Spans()
	if len(spans) != 1 || spans[0].Label != "load plans" {
		t.Errorf("Spans() = %+v, want 1 span", spans)
	}

	recorder.AddError("query timeout")
	if errs := recorder.Errors(); len(errs) != 1 || errs[0] != "query timeout" {
		t.Errorf("Errors() = %+v, want [query timeout]", errs)
	}

	recorder.SetValue("tenant", "acme")
	val, ok := recorder.Value("tenant")
	if !ok || val != "acme" {
		t.Errorf("Value(tenant) = %v, %v, want acme, true", val, ok)
	}
}
