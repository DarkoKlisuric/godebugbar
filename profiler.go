package godebugbar

import (
	"net/http"

	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

type Profiler struct {
	eng *engine.Engine
}

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

func (p *Profiler) Middleware(next http.Handler) http.Handler {
	return p.eng.Middleware(next)
}

func (p *Profiler) Handler() http.Handler {
	return p.eng.Handler()
}

func (p *Profiler) Enabled() bool {
	return p.eng.Enabled()
}
