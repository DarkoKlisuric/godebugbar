package godebugbar

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeCollector(t *testing.T) {
	t.Parallel()

	collector := NewRuntimeCollector()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	section := collector.Collect(t.Context(), req, &ResponseInfo{Status: http.StatusOK})

	if collector.Name() != "runtime" {
		t.Errorf("Name() = %q, want %q", collector.Name(), "runtime")
	}

	if section == nil || section.Toolbar.Label != "Runtime" {
		t.Errorf("section = %+v, want Runtime toolbar item", section)
	}
}
