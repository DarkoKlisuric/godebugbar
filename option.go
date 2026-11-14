package godebugbar

import (
	"net/http"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

type Option func(*engine.Engine)

func WithEnabled(enabled bool) Option {
	return func(e *engine.Engine) { e.SetEnabled(enabled) }
}

func WithAuthorizer(allow func(*http.Request) bool) Option {
	return func(e *engine.Engine) { e.SetAuthorizer(allow) }
}

func WithDashboard(enabled bool) Option {
	return func(e *engine.Engine) { e.SetDashboard(enabled) }
}

func WithServeEndpoints(serve bool) Option {
	return func(e *engine.Engine) { e.SetServeEndpoints(serve) }
}

func WithHTMLInjection() Option {
	return func(e *engine.Engine) { e.EnableHTMLInjection() }
}

func WithMaxHTMLBuffer(bytes int) Option {
	return func(e *engine.Engine) { e.SetMaxBuffer(bytes) }
}

func WithCORS(origins ...string) Option {
	return func(e *engine.Engine) { e.SetCORSOrigins(origins...) }
}

func WithRedactedHeaders(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddHeaders(names...) }
}

func WithRedactedQueryParams(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddQueryParams(names...) }
}

func WithRedactedBodyFields(names ...string) Option {
	return func(e *engine.Engine) { e.Redactor().AddBodyFields(names...) }
}

func WithHiddenSQLValues() Option {
	return func(e *engine.Engine) { e.HideSQLValues() }
}

func WithRequestBody(maxBytes int) Option {
	return func(e *engine.Engine) { e.SetRequestBodyLimit(maxBytes) }
}

func WithStorageSize(size int) Option {
	return func(e *engine.Engine) { e.SetStorageSize(size) }
}

func WithPersistence(path string) Option {
	return func(e *engine.Engine) { e.SetPersistPath(path) }
}

func WithEntryLimit(limit int) Option {
	return func(e *engine.Engine) { e.SetEntryLimit(limit) }
}

func WithPrefix(prefix string) Option {
	return func(e *engine.Engine) { e.SetPrefix(prefix) }
}

func WithPublicURL(url string) Option {
	return func(e *engine.Engine) { e.SetPublicURL(url) }
}

func WithExcludedPrefixes(prefixes ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedPrefixes(prefixes...) }
}

func WithExcludedPaths(paths ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedPaths(paths...) }
}

func WithExcludedMethods(methods ...string) Option {
	return func(e *engine.Engine) { e.AddExcludedMethods(methods...) }
}

func WithRequestCollector() Option {
	return WithCollectors(collector.NewRequestCollector())
}

func WithRuntimeCollector() Option {
	return WithCollectors(collector.NewRuntimeCollector())
}

func WithQueryCollector() Option {
	return WithCollectors(collector.NewQueryCollector())
}

func WithCollectors(collectors ...Collector) Option {
	return func(e *engine.Engine) { e.AddCollectors(collectors...) }
}

func WithoutDefaultCollectors() Option {
	return func(e *engine.Engine) { e.DisableDefaultCollectors() }
}

func WithSlowQueryThreshold(d time.Duration) Option {
	return func(e *engine.Engine) { e.SetSlowQueryThreshold(d) }
}

func WithNPlusOneThreshold(n int) Option {
	return func(e *engine.Engine) { e.SetNPlusOneThreshold(n) }
}

func WithSourceSkip(patterns ...string) Option {
	return func(e *engine.Engine) { e.AddSourceSkip(patterns...) }
}
