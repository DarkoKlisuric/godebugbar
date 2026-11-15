package godebugbar

import (
	"net/http"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

// Option configures a Profiler. New applies options in order.
type Option func(*engine.Engine)

// WithEnabled turns profiling on or off. The default is off.
func WithEnabled(enabled bool) Option {
	return func(e *engine.Engine) { e.SetEnabled(enabled) }
}

// WithAuthorizer limits the profiler UI, API and the X-Debugbar-ID header to
// requests for which allow returns true.
func WithAuthorizer(allow func(*http.Request) bool) Option {
	return func(e *engine.Engine) { e.SetAuthorizer(allow) }
}

// WithDashboard turns the profiler page at the prefix on or off. The default is on.
func WithDashboard(enabled bool) Option {
	return func(e *engine.Engine) { e.SetDashboard(enabled) }
}

// WithServeEndpoints controls whether the middleware serves the profiler
// endpoints. Pass false when Profiler.Handler is mounted elsewhere.
func WithServeEndpoints(serve bool) Option {
	return func(e *engine.Engine) { e.SetServeEndpoints(serve) }
}

// WithHTMLInjection adds the toolbar script to text/html responses.
func WithHTMLInjection() Option {
	return func(e *engine.Engine) { e.EnableHTMLInjection() }
}

// WithMaxHTMLBuffer sets the largest HTML response buffered for toolbar
// injection. Larger responses stream without the toolbar. The default is 4 MiB.
func WithMaxHTMLBuffer(bytes int) Option {
	return func(e *engine.Engine) { e.SetMaxBuffer(bytes) }
}

// WithCORS lets the listed origins call the profiler API and read the
// X-Debugbar-ID header. A "*" wildcard is ignored.
func WithCORS(origins ...string) Option {
	return func(e *engine.Engine) { e.SetCORSOrigins(origins...) }
}

// WithRedactedHeaders masks more request and response headers. Credentials
// such as Authorization and Cookie are masked by default.
func WithRedactedHeaders(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddHeaders(names...) }
}

// WithRedactedQueryParams masks more URL query parameters. Names such as
// token, password and api_key are masked by default.
func WithRedactedQueryParams(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddQueryParams(names...) }
}

// WithRedactedBodyFields masks more JSON or form fields in captured request
// bodies. Names such as password, token and secret are masked by default.
func WithRedactedBodyFields(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddBodyFields(names...) }
}

// WithHiddenSQLValues records SQL with placeholders only, never bound values.
func WithHiddenSQLValues() Option {
	return func(e *engine.Engine) { e.HideSQLValues() }
}

// WithRequestBody captures up to maxBytes of each request body, redacted.
func WithRequestBody(maxBytes int) Option {
	return func(e *engine.Engine) { e.SetRequestBodyLimit(maxBytes) }
}

// WithStorageSize sets how many profiles are kept in memory. The default is 50.
func WithStorageSize(size int) Option {
	return func(e *engine.Engine) { e.SetStorageSize(size) }
}

// WithPersistence keeps profiles in a JSON file at path so they survive restarts.
func WithPersistence(path string) Option {
	return func(e *engine.Engine) { e.SetPersistPath(path) }
}

// WithEntryLimit caps the queries, logs, spans, errors and outgoing calls
// recorded per request. The default is 500 of each.
func WithEntryLimit(limit int) Option {
	return func(e *engine.Engine) { e.SetEntryLimit(limit) }
}

// WithPrefix sets the URL prefix of the profiler. The default is "/_debugbar".
func WithPrefix(prefix string) Option {
	return func(e *engine.Engine) { e.SetPrefix(prefix) }
}

// WithPublicURL sets the absolute profiler URL the browser uses, for an SPA on
// another origin or a separate listener.
func WithPublicURL(url string) Option {
	return func(e *engine.Engine) { e.SetPublicURL(url) }
}

// WithExcludedPrefixes skips profiling for paths with any of the prefixes.
func WithExcludedPrefixes(prefixes ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedPrefixes(prefixes...) }
}

// WithExcludedPaths skips profiling for exact paths.
func WithExcludedPaths(paths ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedPaths(paths...) }
}

// WithExcludedMethods skips profiling for HTTP methods. OPTIONS is always skipped.
func WithExcludedMethods(methods ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedMethods(methods...) }
}

// WithRequestCollector registers the request collector. It is a default
// collector, needed only after WithoutDefaultCollectors.
func WithRequestCollector() Option {
	return WithCollectors(collector.NewRequestCollector())
}

// WithRuntimeCollector registers the Go runtime collector. It is a default
// collector, needed only after WithoutDefaultCollectors.
func WithRuntimeCollector() Option {
	return WithCollectors(collector.NewRuntimeCollector())
}

// WithQueryCollector registers the SQL collector. Database adapters do it for you.
func WithQueryCollector() Option {
	return WithCollectors(collector.NewQueryCollector())
}

// WithCollectors registers custom collectors. New panics when two different
// collectors share a name.
func WithCollectors(collectors ...Collector) Option {
	return func(e *engine.Engine) { e.AddCollectors(collectors...) }
}

// WithoutDefaultCollectors leaves out the default collectors.
func WithoutDefaultCollectors() Option {
	return func(e *engine.Engine) { e.DisableDefaultCollectors() }
}

// WithSlowQueryThreshold marks queries at or above d as slow. The default is 100ms.
func WithSlowQueryThreshold(d time.Duration) Option {
	return func(e *engine.Engine) { e.SetSlowQueryThreshold(d) }
}

// WithNPlusOneThreshold flags a statement run n or more times in one request
// as N+1. The default is 3.
func WithNPlusOneThreshold(n int) Option {
	return func(e *engine.Engine) { e.SetNPlusOneThreshold(n) }
}

// WithSourceSkip skips call frames whose function or file contains a pattern
// when finding the caller of a query.
func WithSourceSkip(patterns ...string) Option {
	return func(e *engine.Engine) { e.AddSourceSkip(patterns...) }
}
