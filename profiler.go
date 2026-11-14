package godebugbar

import (
	"net/http"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

type Profiler struct {
	eng *engine.Engine
}

func New(options ...Option) *Profiler {
	eng := engine.New()
	p := &Profiler{eng: eng}
	engine.Register(p, eng)
	for _, opt := range options {
		if opt != nil {
			opt(p)
		}
	}
	eng.Finalize()
	return p
}

func (p *Profiler) Middleware(next http.Handler) http.Handler {
	return p.eng.Middleware(next)
}
