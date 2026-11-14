package godebugbar

import (
	"context"
	"sync"
	"time"
)

type recorderContextKey struct{}

type QueryEntry struct {
	SQL          string        `json:"sql"`
	Vars         []string      `json:"vars,omitempty"`
	RowsAffected int64         `json:"rows_affected"`
	Duration     time.Duration `json:"duration"`
	Source       string        `json:"source,omitempty"`
	Error        string        `json:"error,omitempty"`
	StartedAt    time.Time     `json:"started_at"`
}

type SpanEntry struct {
	Label    string        `json:"label"`
	Category string        `json:"category"`
	Duration time.Duration `json:"duration"`
	Offset   time.Duration `json:"offset"`
}

type Recorder struct {
	mu        sync.RWMutex
	startedAt time.Time
	route     string
	queries   []QueryEntry
	spans     []SpanEntry
	errors    []string
	values    map[string]any
}

func NewRecorder() *Recorder {
	return &Recorder{
		startedAt: time.Now().UTC(),
		queries:   make([]QueryEntry, 0),
		spans:     make([]SpanEntry, 0),
		errors:    make([]string, 0),
		values:    make(map[string]any),
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

func (r *Recorder) RecordQuery(entry QueryEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queries = append(r.queries, entry)
}

func (r *Recorder) Queries() []QueryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copied := make([]QueryEntry, len(r.queries))
	copy(copied, r.queries)

	return copied
}

func (r *Recorder) StartSpan(label, category string) func() {
	start := time.Now().UTC()
	offset := start.Sub(r.startedAt)

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		r.spans = append(r.spans, SpanEntry{
			Label:    label,
			Category: category,
			Duration: time.Since(start),
			Offset:   offset,
		})
	}
}

func (r *Recorder) Spans() []SpanEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copied := make([]SpanEntry, len(r.spans))
	copy(copied, r.spans)

	return copied
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

	r.errors = append(r.errors, message)
}

func (r *Recorder) Errors() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copied := make([]string, len(r.errors))
	copy(copied, r.errors)

	return copied
}

func (r *Recorder) SetValue(key string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.values[key] = value
}

func (r *Recorder) Value(key string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.values[key]

	return value, ok
}
