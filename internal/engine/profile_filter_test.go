package engine_test

import (
	"net/url"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestProfileFilter(t *testing.T) {
	t.Parallel()

	profile := engine.Profile{
		ID: "abc123", Method: "GET", Path: "/v1/runs/7", Route: "/v1/runs/:id", Status: 500, DurationMs: 42,
		Stats:   engine.ProfileStats{SlowQueries: 1, NPlusOne: 1},
		Queries: []contextstate.QueryEntry{{SQL: "SELECT * FROM group.members WHERE run_id = 7"}},
		Errors:  []string{"timeout talking to llm"},
	}

	cases := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "empty filter", query: "", want: true},
		{name: "path text", query: "q=runs", want: true},
		{name: "route text", query: "q=:id", want: true},
		{name: "sql text", query: "q=group.members", want: true},
		{name: "error text", query: "q=LLM", want: true},
		{name: "id text", query: "q=abc1", want: true},
		{name: "missing text", query: "q=billing", want: false},
		{name: "method match", query: "method=get", want: true},
		{name: "method mismatch", query: "method=POST", want: false},
		{name: "status class", query: "status=5xx", want: true},
		{name: "status class mismatch", query: "status=2", want: false},
		{name: "exact status", query: "status=500", want: true},
		{name: "min duration met", query: "min_ms=40", want: true},
		{name: "min duration not met", query: "min_ms=100", want: false},
		{name: "flags present", query: "has=slow,n_plus_one,error", want: true},
		{name: "flag missing", query: "has=duplicate", want: false},
		{name: "panic missing", query: "has=panic", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			values, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			filter := engine.ParseProfileFilter(values)
			if got := filter.Matches(&profile); got != tc.want {
				t.Fatalf("Matches(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestProfileStats(t *testing.T) {
	t.Parallel()

	if got := engine.NewProfileStats(nil); got != (engine.ProfileStats{}) {
		t.Fatalf("expected zero stats without a recorder, got %+v", got)
	}

	rec := contextstate.NewRecorderWithConfig(contextstate.Config{EntryLimit: 3})
	for i := range 4 {
		rec.RecordQuery(contextstate.QueryEntry{SQL: "SELECT 1", Template: "SELECT ?", Vars: []string{string(rune('a' + i))}})
	}
	rec.AddError("boom")
	rec.AddLog(contextstate.LogEntry{Message: "hi"})
	rec.AddHTTPCall(contextstate.HTTPCallEntry{URL: "https://x.example"})
	rec.SetPanic("p", "stack")

	got := engine.NewProfileStats(rec)
	want := engine.ProfileStats{Queries: 4, NPlusOne: 1, Errors: 1, Logs: 1, HTTPCalls: 1, Panic: true}
	if got != want {
		t.Fatalf("NewProfileStats() = %+v, want %+v", got, want)
	}
}
