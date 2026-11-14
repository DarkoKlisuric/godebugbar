package godebugbar

import "testing"

func TestOptions(t *testing.T) {
	t.Parallel()

	customCollector := &stubCollector{name: "custom"}
	profiler := NewProfiler(
		WithEnabled(false),
		WithStorageSize(10),
		WithPrefix("/_custom"),
		WithCollectors(customCollector),
		WithExcludedPrefixes("/healthz"),
	)

	if profiler.Enabled() {
		t.Error("profiler.Enabled() = true, want false")
	}

	if profiler.Prefix() != "/_custom" {
		t.Errorf("profiler.Prefix() = %q, want %q", profiler.Prefix(), "/_custom")
	}

	if !profiler.IsExcluded("/healthz/ready") {
		t.Error("profiler.IsExcluded(/healthz/ready) = false, want true")
	}
}
