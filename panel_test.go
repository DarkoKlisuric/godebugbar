package godebugbar_test

import (
	"testing"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestPanelBuilder(t *testing.T) {
	t.Parallel()

	p := godebugbar.NewPanel("Custom").
		AddMetrics(godebugbar.NewMetric("Count", "3")).
		AddKeyValue("Section", map[string]string{"a": "b"}).
		AddTable("Rows", []string{"Col"}, [][]string{{"val"}})

	if p.Title != "Custom" || len(p.Widgets) != 3 {
		t.Fatalf("unexpected panel: %+v", p)
	}
}
