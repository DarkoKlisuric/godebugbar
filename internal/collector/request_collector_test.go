package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestRequestCollector(t *testing.T) {
	t.Parallel()

	collector := NewRequestCollector()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/plans?currency=EUR", nil)
	req.Header.Set("Accept", "application/json")

	res := &contract.ResponseInfo{
		Status:   http.StatusOK,
		Size:     256,
		Duration: 1500 * time.Microsecond,
		Headers:  http.Header{"Content-Type": []string{"application/json"}},
		Route:    "/v1/plans",
		Errors:   []string{"warning"},
	}

	section := collector.Collect(context.Background(), req, res)
	if section == nil || section.Toolbar.Color != contract.ColorGreen {
		t.Fatalf("Collect() = %+v, want green section", section)
	}
}
