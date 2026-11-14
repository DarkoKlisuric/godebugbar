package godebugbar

type Option func(*Profiler)

func WithEnabled(enabled bool) Option {
	return func(p *Profiler) {
		p.eng.SetEnabled(enabled)
	}
}

func WithDefaultCollectors(enabled bool) Option {
	return func(p *Profiler) {
		p.eng.SetDefaultCollectors(enabled)
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

func WithCollectors(collectors ...Collector) Option {
	return func(p *Profiler) {
		p.eng.AddCollectors(collectors...)
	}
}

func WithExcludedPrefixes(prefixes ...string) Option {
	return func(p *Profiler) {
		p.eng.AddExcludedPrefixes(prefixes...)
	}
}
