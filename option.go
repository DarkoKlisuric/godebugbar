package godebugbar

import "github.com/DarkoKlisuric/godebugbar/internal/collector"

type Option func(*Profiler)

func WithEnabled(enabled bool) Option {
	return func(p *Profiler) {
		p.eng.SetEnabled(enabled)
	}
}

func WithRequestCollector() Option {
	return func(p *Profiler) {
		p.eng.AddCollectors(collector.NewRequestCollector())
	}
}

func WithQueryCollector() Option {
	return func(p *Profiler) {
		p.eng.AddCollectors(collector.NewQueryCollector())
	}
}

func WithRuntimeCollector() Option {
	return func(p *Profiler) {
		p.eng.AddCollectors(collector.NewRuntimeCollector())
	}
}

func WithCollectors(collectors ...Collector) Option {
	return func(p *Profiler) {
		p.eng.AddCollectors(collectors...)
	}
}

func WithDashboard(enabled bool) Option {
	return func(p *Profiler) {
		p.eng.SetDashboard(enabled)
	}
}

func WithHTMLInjection() Option {
	return func(p *Profiler) {
		p.eng.EnableHTMLInjection()
	}
}

func WithCORS(origins ...string) Option {
	return func(p *Profiler) {
		p.eng.SetCORSOrigins(origins...)
	}
}

func WithStorageSize(size int) Option {
	return func(p *Profiler) {
		p.eng.SetStorageSize(size)
	}
}

func WithPrefix(prefix string) Option {
	return func(p *Profiler) {
		p.eng.SetPrefix(prefix)
	}
}

func WithExcludedPrefixes(prefixes ...string) Option {
	return func(p *Profiler) {
		p.eng.AddExcludedPrefixes(prefixes...)
	}
}

func WithExcludedPaths(paths ...string) Option {
	return func(p *Profiler) {
		p.eng.AddExcludedPaths(paths...)
	}
}

func WithExcludedMethods(methods ...string) Option {
	return func(p *Profiler) {
		p.eng.SetExcludedMethods(methods...)
	}
}
