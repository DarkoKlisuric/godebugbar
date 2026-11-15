package godebugbar

import (
	"net/http"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

// Profiler records a profile for every HTTP request that passes through its
// middleware and serves the profiler UI and JSON API under its prefix.
type Profiler struct {
	eng *engine.Engine
}

// New builds a Profiler. Profiling is off until WithEnabled(true) is passed.
// The request, runtime, log, outgoing HTTP and timeline collectors are on by
// default; WithoutDefaultCollectors turns them off.
func New(options ...Option) *Profiler {
	eng := engine.New()
	for _, opt := range options {
		if opt != nil {
			opt(eng)
		}
	}
	eng.Finalize(
		collector.NewRequestCollector(),
		collector.NewRuntimeCollector(),
		collector.NewLogCollector(),
		collector.NewHTTPCollector(),
		collector.NewTimelineCollector(),
	)

	return &Profiler{eng: eng}
}

// Middleware profiles every request next serves. It returns next unchanged
// when the profiler is disabled.
func (p *Profiler) Middleware(next http.Handler) http.Handler {
	return p.eng.Middleware(next)
}

// Handler serves the profiler UI and API. Mount it on a private listener
// together with WithServeEndpoints(false) and WithPublicURL.
func (p *Profiler) Handler() http.Handler {
	return p.eng.Handler()
}

// Enabled reports whether the profiler records requests.
func (p *Profiler) Enabled() bool {
	return p.eng.Enabled()
}
