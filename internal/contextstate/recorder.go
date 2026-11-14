package contextstate

import (
	"context"
	"sync"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/redact"
)

const (
	DefaultEntryLimit        = 500
	DefaultSlowQuery         = 100 * time.Millisecond
	DefaultNPlusOneThreshold = 3
)

type recorderContextKey struct{}

type Config struct {
	EntryLimit        int
	HideSQLValues     bool
	SourceSkip        []string
	SlowQuery         time.Duration
	Redactor          *redact.Redactor
	NPlusOneThreshold int
}

type Recorder struct {
	mu          sync.RWMutex
	config      Config
	start       time.Time
	route       string
	queries     []QueryEntry
	errors      []string
	logs        []LogEntry
	spans       []SpanEntry
	httpCalls   []HTTPCallEntry
	panicValue  string
	panicStack  string
	requestBody string
	dropped     Dropped
}

type Dropped struct {
	Queries   int `json:"queries,omitempty"`
	Errors    int `json:"errors,omitempty"`
	Logs      int `json:"logs,omitempty"`
	Spans     int `json:"spans,omitempty"`
	HTTPCalls int `json:"http_calls,omitempty"`
}

func NewRecorder() *Recorder {
	return NewRecorderWithConfig(Config{})
}

func NewRecorderWithConfig(config Config) *Recorder {
	if config.EntryLimit <= 0 {
		config.EntryLimit = DefaultEntryLimit
	}
	if config.Redactor == nil {
		config.Redactor = redact.New()
	}
	if config.SlowQuery <= 0 {
		config.SlowQuery = DefaultSlowQuery
	}
	if config.NPlusOneThreshold <= 0 {
		config.NPlusOneThreshold = DefaultNPlusOneThreshold
	}

	return &Recorder{
		config: config,
		start:  time.Now(),
	}
}

func WithRecorder(ctx context.Context, recorder *Recorder) context.Context {
	return context.WithValue(ctx, recorderContextKey{}, recorder)
}

func RecorderFromContext(ctx context.Context) *Recorder {
	if ctx == nil {
		return nil
	}

	recorder, _ := ctx.Value(recorderContextKey{}).(*Recorder)

	return recorder
}

func (r *Recorder) Config() Config {
	return r.config
}

func (r *Recorder) Start() time.Time {
	return r.start
}

func (r *Recorder) RecordQuery(entry QueryEntry) {
	if entry.Template == "" {
		entry.Template = entry.SQL
	}
	if entry.SQL == "" {
		entry.SQL = entry.Template
	}
	if r.config.HideSQLValues {
		entry.SQL = entry.Template
		entry.Vars = nil
	}
	if entry.Source == "" {
		entry.Source = Caller(r.config.SourceSkip)
	}
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now().Add(-entry.Duration)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.queries) >= r.config.EntryLimit {
		r.dropped.Queries++
		return
	}
	r.queries = append(r.queries, entry)
}

func (r *Recorder) Queries() []QueryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneSlice(r.queries)
}

func (r *Recorder) SetRoute(route string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.route = route
}

func (r *Recorder) Route() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.route
}

func (r *Recorder) AddError(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.errors) >= r.config.EntryLimit {
		r.dropped.Errors++
		return
	}
	r.errors = append(r.errors, message)
}

func (r *Recorder) Errors() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneSlice(r.errors)
}

func (r *Recorder) AddLog(entry LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.logs) >= r.config.EntryLimit {
		r.dropped.Logs++
		return
	}
	r.logs = append(r.logs, entry)
}

func (r *Recorder) Logs() []LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneSlice(r.logs)
}

func (r *Recorder) AddSpan(entry SpanEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.spans) >= r.config.EntryLimit {
		r.dropped.Spans++
		return
	}
	r.spans = append(r.spans, entry)
}

func (r *Recorder) Spans() []SpanEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneSlice(r.spans)
}

func (r *Recorder) AddHTTPCall(entry HTTPCallEntry) {
	entry.URL = r.config.Redactor.URL(entry.URL)

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.httpCalls) >= r.config.EntryLimit {
		r.dropped.HTTPCalls++
		return
	}
	r.httpCalls = append(r.httpCalls, entry)
}

func (r *Recorder) HTTPCalls() []HTTPCallEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneSlice(r.httpCalls)
}

func (r *Recorder) SetPanic(value, stack string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.panicValue = value
	r.panicStack = stack
}

func (r *Recorder) Panic() (value, stack string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.panicValue, r.panicStack
}

func (r *Recorder) SetRequestBody(body string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestBody = body
}

func (r *Recorder) RequestBody() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.requestBody
}

func (r *Recorder) Dropped() Dropped {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.dropped
}

func cloneSlice[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)

	return out
}
