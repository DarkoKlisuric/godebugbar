package collector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestQueryCollector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		queries    []contextstate.QueryEntry
		wantColor  contract.BadgeColor
		wantStatus []string
	}{
		{name: "no queries", wantColor: contract.ColorDefault},
		{
			name:       "fast unique query",
			queries:    []contextstate.QueryEntry{{SQL: "SELECT 1", Duration: time.Millisecond}},
			wantColor:  contract.ColorGreen,
			wantStatus: []string{"OK"},
		},
		{
			name: "exact duplicate",
			queries: []contextstate.QueryEntry{
				{SQL: "SELECT * FROM plans WHERE id = 'pro'", Template: "SELECT * FROM plans WHERE id = $1", Vars: []string{"pro"}},
				{SQL: "SELECT * FROM plans WHERE id = 'pro'", Template: "SELECT * FROM plans WHERE id = $1", Vars: []string{"pro"}},
			},
			wantColor:  contract.ColorYellow,
			wantStatus: []string{"DUPLICATE", "DUPLICATE"},
		},
		{
			name: "n plus one with different values",
			queries: []contextstate.QueryEntry{
				{SQL: "SELECT * FROM groups WHERE id = 1", Template: "SELECT * FROM groups WHERE id = $1"},
				{SQL: "SELECT * FROM groups WHERE id = 2", Template: "SELECT * FROM groups WHERE id = $1"},
				{SQL: "SELECT * FROM groups WHERE id = 3", Template: "SELECT * FROM groups WHERE id = $1"},
			},
			wantColor:  contract.ColorYellow,
			wantStatus: []string{"N+1 (3×)", "N+1 (3×)", "N+1 (3×)"},
		},
		{
			name:       "slow query",
			queries:    []contextstate.QueryEntry{{SQL: "SELECT pg_sleep(1)", Duration: time.Second}},
			wantColor:  contract.ColorYellow,
			wantStatus: []string{"SLOW"},
		},
		{
			name:       "failed query wins",
			queries:    []contextstate.QueryEntry{{SQL: "SELECT nope", Error: "column does not exist"}, {SQL: "SELECT pg_sleep(1)", Duration: time.Second}},
			wantColor:  contract.ColorRed,
			wantStatus: []string{"column does not exist", "SLOW"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := contextstate.NewRecorder()
			for _, q := range tc.queries {
				recorder.RecordQuery(q)
			}
			ctx := contextstate.WithRecorder(t.Context(), recorder)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			section := NewQueryCollector().Collect(ctx, req, &contract.ResponseInfo{Status: http.StatusOK})
			if section.Toolbar.Color != tc.wantColor {
				t.Fatalf("color = %q, want %q", section.Toolbar.Color, tc.wantColor)
			}

			table := section.Panel.Widgets[1]
			for i, want := range tc.wantStatus {
				if got := table.Rows[i][6]; !strings.Contains(got, want) {
					t.Errorf("row %d status = %q, want %q", i, got, want)
				}
			}
			if len(table.RowColors) != len(tc.queries) {
				t.Errorf("expected a color per row")
			}
		})
	}
}

func TestQueryCollectorReportsDropped(t *testing.T) {
	t.Parallel()

	recorder := contextstate.NewRecorderWithConfig(contextstate.Config{EntryLimit: 1})
	recorder.RecordQuery(contextstate.QueryEntry{SQL: "SELECT 1"})
	recorder.RecordQuery(contextstate.QueryEntry{SQL: "SELECT 2"})
	ctx := contextstate.WithRecorder(t.Context(), recorder)

	section := NewQueryCollector().Collect(ctx, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil), &contract.ResponseInfo{})
	if !strings.HasPrefix(section.Toolbar.Value, "2 ") {
		t.Fatalf("expected the dropped query in the total, got %q", section.Toolbar.Value)
	}
}
