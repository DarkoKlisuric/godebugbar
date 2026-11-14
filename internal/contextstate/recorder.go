package contextstate

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

type Recorder struct {
	mu      sync.RWMutex
	route   string
	queries []QueryEntry
	errors  []string
}

func NewRecorder() *Recorder {
	return &Recorder{
		queries: make([]QueryEntry, 0),
		errors:  make([]string, 0),
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
