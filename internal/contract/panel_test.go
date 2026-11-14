package contract

import "testing"

func TestPanel(t *testing.T) {
	t.Parallel()

	pairs := map[string]string{"env": "local"}
	panel := NewPanel("Database").
		AddMetrics(NewMetric("Queries", "3")).
		AddKeyValue("Connection", pairs).
		AddTable("Executed Queries", []string{"#", "SQL"}, [][]string{{"1", "SELECT 1"}}).
		AddCode("Plan", "Seq Scan")

	pairs["env"] = "mutated"

	if len(panel.Widgets) != 4 || panel.Widgets[1].Pairs["env"] != "local" {
		t.Errorf("panel.Widgets = %+v, want 4 widgets with copied map", panel.Widgets)
	}
}
