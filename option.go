package godebugbar

type Option func(*Profiler)

func WithEnabled(enabled bool) Option {
	return func(p *Profiler) {
		p.enabled = enabled
	}
}

func WithDefaultCollectors(enabled bool) Option {
	return func(p *Profiler) {
		p.defaultCollectors = enabled
	}
}

func WithStorageSize(size int) Option {
	return func(p *Profiler) {
		p.store = NewStore(size)
	}
}

func WithPrefix(prefix string) Option {
	return func(p *Profiler) {
		if prefix != "" {
			p.prefix = prefix
		}
	}
}

func WithCollectors(collectors ...Collector) Option {
	return func(p *Profiler) {
		for _, collector := range collectors {
			p.AddCollector(collector)
		}
	}
}

func WithExcludedPrefixes(prefixes ...string) Option {
	return func(p *Profiler) {
		p.excludedPrefixes = append(p.excludedPrefixes, prefixes...)
	}
}
