package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

func TestRuntimeCollector(t *testing.T) {
	t.Parallel()

	collector := NewRuntimeCollector()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	section := collector.Collect(t.Context(), req, &contract.ResponseInfo{Status: http.StatusOK})

	if section == nil || section.Toolbar.Label != "Runtime" {
		t.Errorf("section = %+v, want Runtime toolbar item", section)
	}
}
