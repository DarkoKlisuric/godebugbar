package contextstate

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
		SQL:          "SELECT 1",
		RowsAffected: 1,
		Duration:     time.Millisecond,
	})
	if len(recorder.Queries()) != 1 {
		t.Errorf("len(Queries()) = %d, want 1", len(recorder.Queries()))
	}

	recorder.AddError("timeout")
	if len(recorder.Errors()) != 1 {
		t.Errorf("len(Errors()) = %d, want 1", len(recorder.Errors()))
	}
}
