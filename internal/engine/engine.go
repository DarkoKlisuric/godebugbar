package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/redact"
	"github.com/DarkoKlisuric/godebugbar/internal/ui"
)

const (
	DefaultPrefix   = "/_debugbar"
	ProfileHeader   = "X-Debugbar-ID"
	corsAllowAll    = "*"
	explainKeyword1 = "select"
	explainKeyword2 = "with"
)

type Explainer func(ctx context.Context, sql string) ([]string, error)

type Engine struct {
	enabled          bool
	dashboard        bool
	htmlInjection    bool
	serveEndpoints   bool
	defaults         bool
	storageSize      int
	maxBuffer        int
	requestBodyLimit int
	prefix           string
	publicURL        string
	persistPath      string
	corsOrigins      []string
	authorizer       func(*http.Request) bool
	explainer        Explainer
	redactor         *redact.Redactor
	recorderConfig   contextstate.Config
	store            *Store
	collectors       []contract.Collector
	excludedPrefixes []string
	excludedPaths    []string
	excludedMethods  []string
	initializers     []func(*Engine)
}

func New() *Engine {
	redactor := redact.New()

	return &Engine{
		dashboard:       true,
		serveEndpoints:  true,
		defaults:        true,
		storageSize:     defaultStoreCapacity,
		maxBuffer:       DefaultMaxBufferBytes,
		prefix:          DefaultPrefix,
		redactor:        redactor,
		recorderConfig:  contextstate.Config{Redactor: redactor},
		excludedMethods: []string{http.MethodOptions},
	}
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

func (e *Engine) SetServeEndpoints(serve bool) {
	e.serveEndpoints = serve
}

func (e *Engine) DisableDefaultCollectors() {
	e.defaults = false
}

func (e *Engine) SetCORSOrigins(origins ...string) {
	e.corsOrigins = nil
	for _, o := range origins {
		trimmed := strings.TrimRight(strings.TrimSpace(o), "/")
		if trimmed != "" && trimmed != corsAllowAll {
			e.corsOrigins = append(e.corsOrigins, trimmed)
		}
	}
}

func (e *Engine) SetAuthorizer(fn func(*http.Request) bool) {
	e.authorizer = fn
}

func (e *Engine) SetExplainer(fn Explainer) {
	e.explainer = fn
}

func (e *Engine) Redactor() *redact.Redactor {
	return e.redactor
}

func (e *Engine) SetStorageSize(size int) {
	if size > 0 {
		e.storageSize = size
	}
}

func (e *Engine) SetMaxBuffer(size int) {
	if size > 0 {
		e.maxBuffer = size
	}
}

func (e *Engine) MaxBuffer() int {
	return e.maxBuffer
}

func (e *Engine) SetRequestBodyLimit(limit int) {
	e.requestBodyLimit = max(limit, 0)
}

func (e *Engine) SetEntryLimit(limit int) {
	if limit > 0 {
		e.recorderConfig.EntryLimit = limit
	}
}

func (e *Engine) SetSlowQueryThreshold(d time.Duration) {
	if d > 0 {
		e.recorderConfig.SlowQuery = d
	}
}

func (e *Engine) SetNPlusOneThreshold(n int) {
	if n > 1 {
		e.recorderConfig.NPlusOneThreshold = n
	}
}

func (e *Engine) HideSQLValues() {
	e.recorderConfig.HideSQLValues = true
}

func (e *Engine) AddSourceSkip(patterns ...string) {
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p != "" {
			e.recorderConfig.SourceSkip = append(e.recorderConfig.SourceSkip, p)
		}
	}
}

func (e *Engine) SetPersistPath(path string) {
	e.persistPath = strings.TrimSpace(path)
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

func (e *Engine) SetPublicURL(url string) {
	e.publicURL = strings.TrimRight(strings.TrimSpace(url), "/")
}

func (e *Engine) PublicURL() string {
	if e.publicURL != "" {
		return e.publicURL
	}

	return e.prefix
}

func (e *Engine) AddCollectors(collectors ...contract.Collector) {
	for _, c := range collectors {
		if c == nil {
			continue
		}
		idx := slices.IndexFunc(e.collectors, func(existing contract.Collector) bool {
			return existing.Name() == c.Name()
		})
		if idx == -1 {
			e.collectors = append(e.collectors, c)
			continue
		}
		if reflect.TypeOf(e.collectors[idx]) != reflect.TypeOf(c) {
			panic(fmt.Sprintf("godebugbar: two different collectors are named %q; collector names must be unique", c.Name()))
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

func (e *Engine) AddExcludedMethods(methods ...string) {
	for _, m := range methods {
		upper := strings.ToUpper(strings.TrimSpace(m))
		if upper != "" && !slices.Contains(e.excludedMethods, upper) {
			e.excludedMethods = append(e.excludedMethods, upper)
		}
	}
}

func (e *Engine) AddInitializer(fn func(*Engine)) {
	if fn != nil {
		e.initializers = append(e.initializers, fn)
	}
}

func (e *Engine) Finalize(defaults ...contract.Collector) {
	if e.defaults {
		custom := e.collectors
		e.collectors = nil
		e.AddCollectors(defaults...)
		e.AddCollectors(custom...)
	}
	e.store = NewStore(e.storageSize)
	if e.persistPath != "" {
		persister := NewPersister(e.persistPath)
		if profiles, err := persister.Load(); err == nil {
			e.store.Restore(profiles)
		}
		e.store.SetPersister(persister)
	}
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

func (e *Engine) ServesEndpoints() bool {
	return e.serveEndpoints
}

func (e *Engine) Prefix() string {
	return e.prefix
}

func (e *Engine) Store() *Store {
	return e.store
}

func (e *Engine) Collectors() []contract.Collector {
	return slices.Clone(e.collectors)
}

func (e *Engine) Authorized(r *http.Request) bool {
	return e.authorizer == nil || e.authorizer(r)
}

func (e *Engine) IsEndpoint(path string) bool {
	return path == e.prefix || strings.HasPrefix(path, e.prefix+"/")
}

func (e *Engine) Handler() http.Handler {
	mux := http.NewServeMux()
	if e.dashboard {
		mux.HandleFunc("GET "+e.prefix, e.handleDashboard)
		mux.HandleFunc("GET "+e.prefix+"/{$}", e.handleDashboard)
	}
	mux.HandleFunc("GET "+e.prefix+"/assets/toolbar.js", e.handleToolbarAsset)
	mux.HandleFunc("GET "+e.prefix+"/assets/render.js", e.handleRenderAsset)
	mux.HandleFunc("GET "+e.prefix+"/api/profiles", e.handleListProfiles)
	mux.HandleFunc("GET "+e.prefix+"/api/profiles/latest", e.handleLatestProfile)
	mux.HandleFunc("GET "+e.prefix+"/api/profiles/{id}", e.handleGetProfile)
	mux.HandleFunc("GET "+e.prefix+"/api/profiles/{id}/queries/{index}/explain", e.handleExplain)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.applyEndpointCORS(w.Header(), r.Header.Get("Origin"))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !e.Authorized(r) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "godebugbar: not authorized"})
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
		if e.serveEndpoints && e.IsEndpoint(r.URL.Path) {
			apiHandler.ServeHTTP(w, r)
			return
		}
		if e.ShouldSkip(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		session := e.Begin(r)
		rr := NewResponseRecorder(w, r, e.htmlInjection, e.maxBuffer)
		session.PrepareHeaders(rr.Header())

		defer func() {
			if v := recover(); v != nil {
				session.RecordPanic(v)
				status := http.StatusInternalServerError
				if rr.wroteHeader {
					status = rr.StatusCode()
				}
				session.End(e.responseInfo(rr, status))
				panic(v)
			}
		}()
		next.ServeHTTP(rr, session.Request)

		session.End(e.responseInfo(rr, rr.StatusCode()))
		rr.Finish(session.Snippet())
	})
}

func (e *Engine) responseInfo(rr *ResponseRecorder, status int) *contract.ResponseInfo {
	return &contract.ResponseInfo{
		Status:  status,
		Size:    rr.Size(),
		Headers: rr.Header().Clone(),
	}
}

func (e *Engine) RecordRequest(r *http.Request, profileID string, start time.Time, duration time.Duration, res *contract.ResponseInfo) Profile {
	res.Duration = duration
	res.Headers = e.redactor.Header(res.Headers)
	view := e.redactedRequest(r)

	sections := make([]contract.Section, 0, len(e.collectors))
	for _, c := range e.collectors {
		sec := c.Collect(r.Context(), view, res)
		if sec == nil {
			continue
		}
		if sec.Name == "" {
			sec.Name = c.Name()
		}
		sections = append(sections, *sec)
	}

	var queries []contextstate.QueryEntry
	if rec := contextstate.RecorderFromContext(r.Context()); rec != nil {
		queries = rec.Queries()
	}

	prof := Profile{
		ID:          profileID,
		Method:      r.Method,
		Path:        r.URL.Path,
		Route:       res.Route,
		Status:      res.Status,
		Size:        res.Size,
		DurationMs:  float64(duration.Microseconds()) / 1000.0,
		CreatedAt:   start.UTC(),
		Explainable: e.explainer != nil && hasExplainable(queries),
		Sections:    sections,
		Queries:     queries,
	}
	e.store.Save(prof)

	return prof
}

func (e *Engine) redactedRequest(r *http.Request) *http.Request {
	view := r.Clone(r.Context())
	view.Header = e.redactor.Header(r.Header)
	if r.URL.RawQuery != "" {
		u := *r.URL
		u.RawQuery = e.redactor.Query(r.URL.Query()).Encode()
		view.URL = &u
	}

	return view
}

func (e *Engine) NewProfileID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])

	return hex.EncodeToString(b[:])
}

func (e *Engine) Snippet(profileID string) string {
	base := e.PublicURL()

	return fmt.Sprintf(
		`<script src="%s/assets/toolbar.js" data-base-url="%s" data-profile-id="%s" defer></script>`,
		base,
		base,
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
	if !strings.Contains(strings.ToLower(existing), strings.ToLower(ProfileHeader)) {
		h.Set("Access-Control-Expose-Headers", existing+", "+ProfileHeader)
	}
}

func (e *Engine) applyEndpointCORS(h http.Header, reqOrigin string) {
	if len(e.corsOrigins) == 0 {
		return
	}
	h.Add("Vary", "Origin")
	if reqOrigin == "" || !slices.Contains(e.corsOrigins, reqOrigin) {
		return
	}
	h.Set("Access-Control-Allow-Origin", reqOrigin)
	h.Set("Access-Control-Allow-Credentials", "true")
	h.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type")
}

func (e *Engine) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ui.DashboardHTML(e.PublicURL())))
}

func (e *Engine) handleToolbarAsset(w http.ResponseWriter, _ *http.Request) {
	writeScript(w, ui.ToolbarScript())
}

func (e *Engine) handleRenderAsset(w http.ResponseWriter, _ *http.Request) {
	writeScript(w, ui.RenderScript())
}

func (e *Engine) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, e.store.Summaries(r.URL.Query().Get("since")))
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
	prof, ok := e.store.Get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile not found"})
		return
	}
	writeJSON(w, http.StatusOK, prof)
}

func (e *Engine) handleExplain(w http.ResponseWriter, r *http.Request) {
	if e.explainer == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no explainer registered"})
		return
	}
	prof, ok := e.store.Get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile not found"})
		return
	}
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || index < 1 || index > len(prof.Queries) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "query not found"})
		return
	}
	query := prof.Queries[index-1]
	if !isExplainable(query.SQL) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only SELECT statements can be explained"})
		return
	}

	plan, err := e.explainer(r.Context(), query.SQL)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sql": query.SQL, "plan": plan})
}

func hasExplainable(queries []contextstate.QueryEntry) bool {
	return slices.ContainsFunc(queries, func(q contextstate.QueryEntry) bool {
		return isExplainable(q.SQL)
	})
}

func isExplainable(sql string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(sql))
	if strings.Contains(trimmed, ";") {
		return false
	}

	return strings.HasPrefix(trimmed, explainKeyword1) || strings.HasPrefix(trimmed, explainKeyword2)
}

func writeScript(w http.ResponseWriter, script string) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(script))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
