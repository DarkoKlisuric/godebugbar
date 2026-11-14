package contextstate

import (
	"context"
	"strings"
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
	recorder.RecordQuery(QueryEntry{SQL: "SELECT 1", RowsAffected: 1, Duration: time.Millisecond})
	recorder.AddError("timeout")
	recorder.AddLog(LogEntry{Level: "INFO", Message: "hi"})
	recorder.AddSpan(SpanEntry{Name: "work"})
	recorder.AddHTTPCall(HTTPCallEntry{Method: "GET", URL: "https://x.example/?token=abc"})
	recorder.SetPanic("boom", "stack")
	recorder.SetRequestBody("{}")

	if recorder.Route() != "/v1/plans" || len(recorder.Queries()) != 1 || len(recorder.Errors()) != 1 ||
		len(recorder.Logs()) != 1 || len(recorder.Spans()) != 1 || len(recorder.HTTPCalls()) != 1 || recorder.RequestBody() != "{}" {
		t.Fatalf("expected one of each entry")
	}
	if value, stack := recorder.Panic(); value != "boom" || stack != "stack" {
		t.Fatalf("Panic() = %q %q", value, stack)
	}
	if strings.Contains(recorder.HTTPCalls()[0].URL, "abc") {
		t.Fatalf("expected the outgoing URL to be redacted, got %s", recorder.HTTPCalls()[0].URL)
	}

	q := recorder.Queries()[0]
	if q.Template != "SELECT 1" || q.Source == "" || q.StartedAt.IsZero() {
		t.Fatalf("expected defaults for template, source and start, got %+v", q)
	}
	if !strings.Contains(q.Source, "recorder_test.go") {
		t.Fatalf("expected the caller to be this test file, got %q", q.Source)
	}
}

func TestRecorderLimits(t *testing.T) {
	t.Parallel()

	recorder := NewRecorderWithConfig(Config{EntryLimit: 2})
	for range 5 {
		recorder.RecordQuery(QueryEntry{SQL: "SELECT 1"})
		recorder.AddError("e")
		recorder.AddLog(LogEntry{})
		recorder.AddSpan(SpanEntry{})
		recorder.AddHTTPCall(HTTPCallEntry{})
	}

	want := Dropped{Queries: 3, Errors: 3, Logs: 3, Spans: 3, HTTPCalls: 3}
	if got := recorder.Dropped(); got != want {
		t.Fatalf("Dropped() = %+v, want %+v", got, want)
	}
	if len(recorder.Queries()) != 2 {
		t.Fatalf("expected the limit to cap stored queries")
	}
}

func TestRecorderHidesSQLValues(t *testing.T) {
	t.Parallel()

	recorder := NewRecorderWithConfig(Config{HideSQLValues: true})
	recorder.RecordQuery(QueryEntry{SQL: "SELECT 'secret'", Template: "SELECT ?", Vars: []string{"secret"}})

	q := recorder.Queries()[0]
	if q.SQL != "SELECT ?" || q.Vars != nil {
		t.Fatalf("expected values to be dropped, got %+v", q)
	}
}

func TestCallerSkip(t *testing.T) {
	t.Parallel()

	if got := Caller(nil); !strings.Contains(got, "recorder_test.go") {
		t.Fatalf("Caller() = %q, want this file", got)
	}
	if got := Caller([]string{"recorder_test.go"}); strings.Contains(got, "recorder_test.go") {
		t.Fatalf("expected the skip pattern to skip this file, got %q", got)
	}
}
