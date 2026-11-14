package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestOptionalCollectorsHideWhenEmpty(t *testing.T) {
	t.Parallel()

	collectors := []contract.Collector{NewLogCollector(), NewHTTPCollector(), NewTimelineCollector()}
	for _, c := range collectors {
		t.Run(c.Name(), func(t *testing.T) {
			t.Parallel()

			ctx := contextstate.WithRecorder(t.Context(), contextstate.NewRecorder())
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
			if sec := c.Collect(ctx, req, &contract.ResponseInfo{}); sec != nil {
				t.Fatalf("expected nil for an empty request, got %+v", sec)
			}
			if sec := c.Collect(t.Context(), req, &contract.ResponseInfo{}); sec != nil {
				t.Fatalf("expected nil without a recorder")
			}
		})
	}
}

func TestLogHTTPAndTimelineCollectors(t *testing.T) {
	t.Parallel()

	recorder := contextstate.NewRecorder()
	now := time.Now()
	recorder.AddLog(contextstate.LogEntry{Time: now, Level: "WARN", Message: "slow", Attrs: map[string]string{"b": "2", "a": "1"}})
	recorder.AddHTTPCall(contextstate.HTTPCallEntry{Method: "GET", URL: "https://api.example", Status: http.StatusBadGateway, StartedAt: now, Duration: time.Millisecond})
	recorder.AddSpan(contextstate.SpanEntry{Name: "render", StartedAt: now, Duration: time.Millisecond})
	recorder.RecordQuery(contextstate.QueryEntry{SQL: "SELECT 1", StartedAt: now, Duration: time.Millisecond})
	ctx := contextstate.WithRecorder(t.Context(), recorder)
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/x", nil)
	res := &contract.ResponseInfo{Status: http.StatusOK, Duration: 5 * time.Millisecond}

	cases := []struct {
		name      string
		collector contract.Collector
		wantColor contract.BadgeColor
		check     func(*testing.T, *contract.Section)
	}{
		{
			name: "logs", collector: NewLogCollector(), wantColor: contract.ColorYellow,
			check: func(t *testing.T, s *contract.Section) {
				t.Helper()
				if got := s.Panel.Widgets[1].Rows[0][3]; got != "a=1 b=2" {
					t.Errorf("attrs = %q, want sorted", got)
				}
			},
		},
		{name: "http", collector: NewHTTPCollector(), wantColor: contract.ColorYellow},
		{
			name: "timeline", collector: NewTimelineCollector(), wantColor: contract.ColorDefault,
			check: func(t *testing.T, s *contract.Section) {
				t.Helper()
				events := s.Panel.Widgets[0].Events
				if len(events) != 4 || events[0].Category != "request" {
					t.Errorf("expected request + 3 events, got %+v", events)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sec := tc.collector.Collect(ctx, req, res)
			if sec == nil || sec.Toolbar.Color != tc.wantColor {
				t.Fatalf("section = %+v, want color %q", sec, tc.wantColor)
			}
			if tc.check != nil {
				tc.check(t, sec)
			}
		})
	}
}
