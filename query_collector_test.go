package godebugbar

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryCollector(t *testing.T) {
	t.Parallel()

	recorder := NewRecorder()
	recorder.RecordQuery(QueryEntry{
		SQL:          "SELECT * FROM billing.plans WHERE id = $1",
		Vars:         []string{"pro"},
		RowsAffected: 1,
		Duration:     500 * time.Microsecond,
		Source:       "plan_repository.go:42",
	})
	recorder.RecordQuery(QueryEntry{
		SQL:          "SELECT * FROM billing.plans WHERE id = $1",
		Vars:         []string{"pro"},
		RowsAffected: 1,
		Duration:     400 * time.Microsecond,
		Source:       "plan_repository.go:42",
	})

	ctx := WithRecorder(t.Context(), recorder)
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/plans", nil)

	collector := NewQueryCollector()
	section := collector.Collect(ctx, req, &ResponseInfo{Status: http.StatusOK})

	if collector.Name() != "database" {
		t.Errorf("Name() = %q, want %q", collector.Name(), "database")
	}

	if section.Toolbar.Color != ColorYellow {
		t.Errorf("Toolbar.Color = %q, want %q for duplicate query", section.Toolbar.Color, ColorYellow)
	}

	if section.Toolbar.Tooltip["Duplicates"] != "1" {
		t.Errorf("Tooltip[Duplicates] = %q, want 1", section.Toolbar.Tooltip["Duplicates"])
	}
}
