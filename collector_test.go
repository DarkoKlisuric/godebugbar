package godebugbar_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar"
)

type stubCollector struct{}

func (c *stubCollector) Name() string {
	return "stub"
}

func (c *stubCollector) Collect(_ context.Context, _ *http.Request, res *godebugbar.ResponseInfo) *godebugbar.Section {
	return godebugbar.NewSection(godebugbar.ToolbarItem{
		Label: "Stub",
		Value: "OK",
		Color: godebugbar.StatusColor(res.Status),
	}, godebugbar.NewPanel("Stub"))
}

func TestCollectorContract(t *testing.T) {
	t.Parallel()

	var c godebugbar.Collector = &stubCollector{}
	sec := c.Collect(t.Context(), nil, &godebugbar.ResponseInfo{Status: http.StatusOK, Duration: time.Millisecond})
	if c.Name() != "stub" || sec.Toolbar.Color != godebugbar.ColorGreen {
		t.Fatalf("unexpected section: %+v", sec)
	}
}
