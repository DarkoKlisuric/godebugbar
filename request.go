package godebugbar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
)

// Request records data into the profile of the current HTTP request. All
// methods are no-ops on a nil *Request.
type Request struct {
	rec *contextstate.Recorder
}

// FromContext returns the recorder of the request ctx belongs to, or nil
// when the request is not profiled.
func FromContext(ctx context.Context) *Request {
	rec := contextstate.RecorderFromContext(ctx)
	if rec == nil {
		return nil
	}

	return &Request{rec: rec}
}

// AddQuery records a database query. Database adapters call it for you.
func (r *Request) AddQuery(q Query) {
	if r == nil {
		return
	}
	r.rec.RecordQuery(q.entry())
}

// AddError records an error.
func (r *Request) AddError(err error) {
	if r == nil || err == nil {
		return
	}
	r.rec.AddError(err.Error())
}

// Log records a log message. NewSlogHandler does this for slog records.
func (r *Request) Log(level slog.Level, msg string, attrs ...slog.Attr) {
	if r == nil {
		return
	}
	values := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		values[attr.Key] = attr.Value.String()
	}
	r.rec.AddLog(contextstate.LogEntry{
		Time:    time.Now(),
		Level:   level.String(),
		Message: msg,
		Attrs:   values,
	})
}

// StartSpan starts a timeline span. Call the returned function to end it.
func (r *Request) StartSpan(name string) func() {
	if r == nil {
		return func() {}
	}
	start := time.Now()

	return func() {
		r.rec.AddSpan(contextstate.SpanEntry{Name: name, StartedAt: start, Duration: time.Since(start)})
	}
}

// SetRoute records the matched route pattern, such as "/users/{id}".
func (r *Request) SetRoute(route string) {
	if r == nil {
		return
	}
	r.rec.SetRoute(route)
}

func formatArgs(args []any) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	for i, arg := range args {
		out[i] = fmt.Sprint(arg)
	}

	return out
}
