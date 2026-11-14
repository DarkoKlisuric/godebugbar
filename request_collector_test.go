package godebugbar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestCollector(t *testing.T) {
	t.Parallel()

	collector := NewRequestCollector()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/plans?currency=EUR", nil)
	req.Header.Set("Accept", "application/json")

	res := &ResponseInfo{
		Status:   http.StatusOK,
		Size:     256,
		Duration: 1500 * time.Microsecond,
		Headers:  http.Header{"Content-Type": []string{"application/json"}},
		Route:    "/v1/plans",
		Errors:   []string{"minor warning"},
	}

	section := collector.Collect(context.Background(), req, res)
	if section == nil {
		t.Fatal("Collect() = nil, want section")
	}

	if collector.Name() != "request" {
		t.Errorf("Name() = %q, want %q", collector.Name(), "request")
	}

	if section.Toolbar.Color != ColorGreen {
		t.Errorf("Toolbar.Color = %q, want %q", section.Toolbar.Color, ColorGreen)
	}

	if len(section.Panel.Widgets) != 5 {
		t.Errorf("len(Panel.Widgets) = %d, want 5", len(section.Panel.Widgets))
	}
}
