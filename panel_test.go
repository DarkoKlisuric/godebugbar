package godebugbar

import "testing"

func TestPanel(t *testing.T) {
	t.Parallel()

	pairs := map[string]string{"env": "local"}
	panel := NewPanel("Database").
		AddMetrics(NewMetric("Queries", "3"), NewMetric("Duration", "1.4 ms")).
		AddKeyValue("Connection", pairs).
		AddTable("Executed Queries", []string{"#", "SQL"}, [][]string{{"1", "SELECT 1"}}).
		AddCode("Raw Plan", "Seq Scan")

	pairs["env"] = "mutated"

	if panel.Title != "Database" {
		t.Fatalf("panel.Title = %q, want %q", panel.Title, "Database")
	}

	if len(panel.Widgets) != 4 {
		t.Fatalf("len(panel.Widgets) = %d, want 4", len(panel.Widgets))
	}

	if panel.Widgets[0].Kind != WidgetKindMetrics || len(panel.Widgets[0].Metrics) != 2 {
		t.Errorf("widget[0] = %+v, want 2 metrics", panel.Widgets[0])
	}

	if panel.Widgets[1].Kind != WidgetKindKeyValue || panel.Widgets[1].Pairs["env"] != "local" {
		t.Errorf("widget[1] pairs = %+v, want copied map with env=local", panel.Widgets[1].Pairs)
	}

	if panel.Widgets[2].Kind != WidgetKindTable || len(panel.Widgets[2].Rows) != 1 {
		t.Errorf("widget[2] = %+v, want table with 1 row", panel.Widgets[2])
	}

	if panel.Widgets[3].Kind != WidgetKindCode || panel.Widgets[3].Code != "Seq Scan" {
		t.Errorf("widget[3] = %+v, want code widget", panel.Widgets[3])
	}
}
