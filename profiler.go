package godebugbar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ProfileHeader      = "X-Debugbar-ID"
	defaultPrefix      = "/_debugbar"
	profilesSuffix     = "/api/profiles"
	toolbarScriptFmt   = `<script src="%s/toolbar.js" data-profile-id="%s" data-prefix="%s" defer></script>`
	exposeHeaderName   = "Access-Control-Expose-Headers"
	allowOriginHeader  = "Access-Control-Allow-Origin"
	jsonContentTypeVal = "application/json; charset=utf-8"
)

type Profiler struct {
	mu               sync.RWMutex
	enabled          bool
	prefix           string
	store            *Store
	collectors       []Collector
	excludedPrefixes []string
	sequence         atomic.Uint64
}

func NewProfiler(options ...Option) *Profiler {
	profiler := &Profiler{
		enabled:    true,
		prefix:     defaultPrefix,
		store:      NewStore(defaultStoreCapacity),
		collectors: make([]Collector, 0),
	}

	for _, option := range options {
		option(profiler)
	}

	return profiler
}

func (p *Profiler) Enabled() bool {
	return p.enabled
}

func (p *Profiler) Prefix() string {
	return p.prefix
}

func (p *Profiler) Store() *Store {
	return p.store
}

func (p *Profiler) AddCollector(collector Collector) {
	if collector == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.collectors = append(p.collectors, collector)
}

func (p *Profiler) IsExcluded(path string) bool {
	if path == p.prefix || strings.HasPrefix(path, p.prefix+"/") {
		return true
	}

	for _, excluded := range p.excludedPrefixes {
		if excluded != "" && strings.HasPrefix(path, excluded) {
			return true
		}
	}

	return false
}

func (p *Profiler) NewProfileID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		seq := p.sequence.Add(1)

		return fmt.Sprintf("dbg-%d", seq)
	}

	return hex.EncodeToString(raw[:])
}

func (p *Profiler) Snippet(profileID string) string {
	return fmt.Sprintf(toolbarScriptFmt, p.prefix, profileID, p.prefix)
}

func (p *Profiler) Middleware(next http.Handler) http.Handler {
	debugHandler := p.Handler()

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !p.enabled {
			next.ServeHTTP(writer, request)

			return
		}

		if request.URL.Path == p.prefix || strings.HasPrefix(request.URL.Path, p.prefix+"/") {
			debugHandler.ServeHTTP(writer, request)

			return
		}

		if p.IsExcluded(request.URL.Path) {
			next.ServeHTTP(writer, request)

			return
		}

		profileID := p.NewProfileID()
		startedAt := time.Now().UTC()
		recorder := NewRecorder()
		ctx := WithRecorder(request.Context(), recorder)

		wrappedWriter := newResponseRecorder(writer)
		wrappedWriter.Header().Set(ProfileHeader, profileID)
		wrappedWriter.Header().Add(exposeHeaderName, ProfileHeader)

		next.ServeHTTP(wrappedWriter, request.WithContext(ctx))

		route := recorder.Route()
		if route == "" {
			route = request.URL.Path
		}

		responseInfo := &ResponseInfo{
			Status:   wrappedWriter.status,
			Size:     wrappedWriter.size,
			Duration: time.Since(startedAt),
			Headers:  wrappedWriter.Header().Clone(),
			Route:    route,
			Errors:   recorder.Errors(),
		}

		p.RecordRequest(ctx, profileID, startedAt, request, responseInfo)
		wrappedWriter.FlushBuffered(p.Snippet(profileID))
	})
}

func (p *Profiler) RecordRequest(ctx context.Context, profileID string, startedAt time.Time, request *http.Request, response *ResponseInfo) *Profile {
	p.mu.RLock()
	collectors := make([]Collector, len(p.collectors))
	copy(collectors, p.collectors)
	p.mu.RUnlock()

	sections := make([]*Section, 0, len(collectors))
	for _, collector := range collectors {
		section := collector.Collect(ctx, request, response)
		if section == nil {
			continue
		}

		section.Name = collector.Name()
		sections = append(sections, section)
	}

	route := response.Route
	if route == "" {
		route = request.URL.Path
	}

	profile := &Profile{
		ID:         profileID,
		Method:     request.Method,
		Path:       request.URL.Path,
		Route:      route,
		Status:     response.Status,
		DurationMs: float64(response.Duration.Microseconds()) / 1000.0,
		Size:       response.Size,
		StartedAt:  startedAt,
		Sections:   sections,
	}

	p.store.Save(profile)

	return profile
}

func (p *Profiler) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(allowOriginHeader, "*")

		trimmed := strings.TrimPrefix(request.URL.Path, p.prefix)
		if trimmed == profilesSuffix {
			p.writeJSON(writer, http.StatusOK, p.store.List())

			return
		}

		if after, ok := strings.CutPrefix(trimmed, profilesSuffix+"/"); ok {
			id := after
			profile, ok := p.store.Get(id)
			if !ok {
				http.NotFound(writer, request)

				return
			}

			p.writeJSON(writer, http.StatusOK, profile)

			return
		}

		http.NotFound(writer, request)
	})
}

func (p *Profiler) writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set(contentTypeHeader, jsonContentTypeVal)
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
