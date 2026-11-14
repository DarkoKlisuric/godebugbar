package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestQueryCollector(t *testing.T) {
	t.Parallel()

	recorder := contextstate.NewRecorder()
	recorder.RecordQuery(contextstate.QueryEntry{
		SQL:          "SELECT * FROM billing.plans WHERE id = $1",
		Vars:         []string{"pro"},
		RowsAffected: 1,
		Duration:     500 * time.Microsecond,
	})
	recorder.RecordQuery(contextstate.QueryEntry{
		SQL:          "SELECT * FROM billing.plans WHERE id = $1",
		Vars:         []string{"pro"},
		RowsAffected: 1,
		Duration:     400 * time.Microsecond,
	})

	ctx := contextstate.WithRecorder(t.Context(), recorder)
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/plans", nil)

	collector := NewQueryCollector()
	section := collector.Collect(ctx, req, &contract.ResponseInfo{Status: http.StatusOK})

	if section.Toolbar.Color != contract.ColorYellow {
		t.Errorf("Toolbar.Color = %q, want %q for duplicate query", section.Toolbar.Color, contract.ColorYellow)
	}
}
