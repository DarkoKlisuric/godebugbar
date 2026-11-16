package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestRequestCollector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		res         contract.ResponseInfo
		body        string
		wantColor   contract.BadgeColor
		wantWidgets []string
	}{
		{
			name:        "ok with errors and query",
			res:         contract.ResponseInfo{Status: http.StatusOK, Route: "/v1/plans", Errors: []string{"warning"}},
			wantColor:   contract.ColorGreen,
			wantWidgets: []string{"Request Headers", "Response Headers", "Query Parameters", "Errors"},
		},
		{
			name:        "panic is red with its stack",
			res:         contract.ResponseInfo{Status: http.StatusInternalServerError, Panic: "boom", PanicStack: "goroutine 1"},
			wantColor:   contract.ColorRed,
			wantWidgets: []string{"Panic: boom"},
		},
		{
			name:        "captured body",
			res:         contract.ResponseInfo{Status: http.StatusCreated},
			body:        `{"name":"group"}`,
			wantColor:   contract.ColorGreen,
			wantWidgets: []string{"Request Body"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := contextstate.NewRecorder()
			recorder.SetRequestBody(tc.body)
			ctx := contextstate.WithRecorder(t.Context(), recorder)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/plans?currency=EUR", nil)
			res := tc.res
			res.Duration = 1500 * time.Microsecond

			section := NewRequestCollector().Collect(ctx, req, &res)
			if section.Toolbar.Color != tc.wantColor {
				t.Fatalf("color = %q, want %q", section.Toolbar.Color, tc.wantColor)
			}
			titles := map[string]bool{}
			for _, w := range section.Panel.Widgets {
				titles[w.Title] = true
			}
			for _, want := range tc.wantWidgets {
				if !titles[want] {
					t.Errorf("missing widget %q in %v", want, titles)
				}
			}
		})
	}
}
