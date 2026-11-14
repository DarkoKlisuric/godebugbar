package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/ui"
)

const (
	DefaultPrefix = "/_debugbar"
	ProfileHeader = "X-Debugbar-ID"
)

var registry sync.Map

type Engine struct {
	enabled          bool
	dashboard        bool
	htmlInjection    bool
	storageSize      int
	prefix           string
	corsOrigins      []string
	store            *Store
	collectors       []contract.Collector
	excludedPrefixes []string
	excludedPaths    []string
	excludedMethods  []string
	initializers     []func(*Engine)
}

func New() *Engine {
	return &Engine{
		enabled:         true,
		dashboard:       true,
		storageSize:     defaultStoreCapacity,
		prefix:          DefaultPrefix,
		excludedMethods: []string{http.MethodOptions},
	}
}

func Register(owner any, eng *Engine) {
	registry.Store(owner, eng)
}

func Lookup(owner any) *Engine {
	v, ok := registry.Load(owner)
	if !ok {
		return nil
	}
	eng, _ := v.(*Engine)
	return eng
}

func (e *Engine) SetEnabled(enabled bool) {
	e.enabled = enabled
}

func (e *Engine) SetDashboard(enabled bool) {
	e.dashboard = enabled
}

func (e *Engine) EnableHTMLInjection() {
	e.htmlInjection = true
}

func (e *Engine) SetCORSOrigins(origins ...string) {
	e.corsOrigins = nil
	for _, o := range origins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			e.corsOrigins = append(e.corsOrigins, trimmed)
		}
	}
}

func (e *Engine) SetStorageSize(size int) {
	if size > 0 {
		e.storageSize = size
	}
}

func (e *Engine) SetPrefix(prefix string) {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	e.prefix = strings.TrimRight(trimmed, "/")
}

func (e *Engine) AddCollectors(collectors ...contract.Collector) {
	for _, c := range collectors {
		if c == nil {
			continue
		}
		exists := slices.ContainsFunc(e.collectors, func(existing contract.Collector) bool {
			return existing.Name() == c.Name()
		})
		if !exists {
			e.collectors = append(e.collectors, c)
		}
	}
}

func (e *Engine) AddExcludedPrefixes(prefixes ...string) {
	for _, pfx := range prefixes {
		if pfx != "" {
			e.excludedPrefixes = append(e.excludedPrefixes, pfx)
		}
	}
}

func (e *Engine) AddExcludedPaths(paths ...string) {
	for _, p := range paths {
		if p != "" {
			e.excludedPaths = append(e.excludedPaths, p)
		}
	}
}

func (e *Engine) SetExcludedMethods(methods ...string) {
	e.excludedMethods = nil
	for _, m := range methods {
		upper := strings.ToUpper(strings.TrimSpace(m))
		if upper != "" {
			e.excludedMethods = append(e.excludedMethods, upper)
		}
	}
}

func (e *Engine) AddInitializer(fn func(*Engine)) {
	if fn != nil {
		e.initializers = append(e.initializers, fn)
	}
}

func (e *Engine) Finalize() {
	e.store = NewStore(e.storageSize)
	for _, fn := range e.initializers {
		fn(e)
	}
}

func (e *Engine) Enabled() bool {
	return e.enabled
}

func (e *Engine) DashboardEnabled() bool {
	return e.dashboard
}

func (e *Engine) HTMLInjectionEnabled() bool {
	return e.htmlInjection
}

func (e *Engine) Prefix() string {
	return e.prefix
}

func (e *Engine) Store() *Store {
	return e.store
}

func (e *Engine) Handler() http.Handler {
	mux := http.NewServeMux()
	if e.dashboard {
		mux.HandleFunc(e.prefix, e.handleDashboard)
		mux.HandleFunc(e.prefix+"/", e.handleDashboard)
	}
	mux.HandleFunc(e.prefix+"/assets/toolbar.js", e.handleToolbarAsset)
	mux.HandleFunc(e.prefix+"/api/profiles/latest", e.handleLatestProfile)
	mux.HandleFunc(e.prefix+"/api/profiles/", e.handleGetProfile)
	mux.HandleFunc(e.prefix+"/api/profiles", e.handleListProfiles)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.applyEndpointCORS(w.Header(), r.Header.Get("Origin"))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (e *Engine) Middleware(next http.Handler) http.Handler {
	if !e.enabled {
		return next
	}

	apiHandler := e.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, e.prefix+"/") || r.URL.Path == e.prefix {
			apiHandler.ServeHTTP(w, r)
			return
		}

		if e.ShouldSkip(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		recorder := contextstate.NewRecorder()
		ctx := contextstate.WithRecorder(r.Context(), recorder)
		req := r.WithContext(ctx)
		rr := NewResponseRecorder(w)

		profileID := e.NewProfileID()
		rr.Header().Set(ProfileHeader, profileID)
		e.ExposeProfileHeader(rr.Header())

		start := time.Now()
		next.ServeHTTP(rr, req)
		duration := time.Since(start)

		e.RecordRequest(req, profileID, start, duration, &contract.ResponseInfo{
			Status:   rr.StatusCode(),
			Size:     rr.Size(),
			Duration: duration,
			Headers:  rr.Header().Clone(),
			Route:    recorder.Route(),
			Errors:   recorder.Errors(),
		})

		snippet := ""
		if e.htmlInjection {
			snippet = e.Snippet(profileID)
		}
		rr.FlushWithSnippet(snippet)
	})
}

func (e *Engine) RecordRequest(r *http.Request, profileID string, start time.Time, duration time.Duration, res *contract.ResponseInfo) Profile {
	sections := make([]contract.Section, 0, len(e.collectors))
	for _, c := range e.collectors {
		sec := c.Collect(r.Context(), r, res)
		if sec == nil {
			continue
		}
		if sec.Name == "" {
			sec.Name = c.Name()
		}
		sections = append(sections, *sec)
	}

	prof := Profile{
		ID:         profileID,
		Method:     r.Method,
		Path:       r.URL.Path,
		Route:      res.Route,
		Status:     res.Status,
		Size:       res.Size,
		DurationMs: float64(duration.Microseconds()) / 1000.0,
		CreatedAt:  start.UTC(),
		Sections:   sections,
	}
	e.store.Save(prof)
	return prof
}

func (e *Engine) NewProfileID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (e *Engine) Snippet(profileID string) string {
	return fmt.Sprintf(
		`<script src="%s/assets/toolbar.js" data-base-url="%s" data-profile-id="%s" defer></script>`,
		e.prefix,
		e.prefix,
		profileID,
	)
}

func (e *Engine) ShouldSkip(method, path string) bool {
	if slices.Contains(e.excludedMethods, strings.ToUpper(method)) {
		return true
	}
	return e.IsExcluded(path)
}

func (e *Engine) IsExcluded(path string) bool {
	if slices.Contains(e.excludedPaths, path) {
		return true
	}
	for _, pfx := range e.excludedPrefixes {
		if strings.HasPrefix(path, pfx) {
			return true
		}
	}
	return false
}

func (e *Engine) ExposeProfileHeader(h http.Header) {
	if len(e.corsOrigins) == 0 {
		return
	}
	existing := h.Get("Access-Control-Expose-Headers")
	if existing == "" {
		h.Set("Access-Control-Expose-Headers", ProfileHeader)
		return
	}
	if !strings.Contains(existing, ProfileHeader) {
		h.Set("Access-Control-Expose-Headers", existing+", "+ProfileHeader)
	}
}

func (e *Engine) applyEndpointCORS(h http.Header, reqOrigin string) {
	if len(e.corsOrigins) == 0 {
		return
	}
	allowed := ""
	if slices.Contains(e.corsOrigins, "*") {
		allowed = "*"
	} else if reqOrigin != "" && slices.Contains(e.corsOrigins, reqOrigin) {
		allowed = reqOrigin
	}
	if allowed != "" {
		h.Set("Access-Control-Allow-Origin", allowed)
		h.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

func (e *Engine) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != e.prefix && r.URL.Path != e.prefix+"/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ui.DashboardHTML(e.prefix)))
}

func (e *Engine) handleToolbarAsset(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ui.ToolbarScript()))
}

func (e *Engine) handleListProfiles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, e.store.List())
}

func (e *Engine) handleLatestProfile(w http.ResponseWriter, _ *http.Request) {
	prof, ok := e.store.Latest()
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no profiles recorded"})
		return
	}
	writeJSON(w, http.StatusOK, prof)
}

func (e *Engine) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, e.prefix+"/api/profiles/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing profile id"})
		return
	}
	prof, ok := e.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile not found"})
		return
	}
	writeJSON(w, http.StatusOK, prof)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
