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

func TestPanelColoredTableAndTimeline(t *testing.T) {
	t.Parallel()

	p := NewPanel("x").
		AddColoredTable("t", []string{"a"}, [][]string{{"1"}}, []BadgeColor{ColorRed}).
		AddTimeline("tl", []TimelineEvent{{Label: "e", DurationMs: 1}})

	if p.Widgets[0].Kind != WidgetKindTable || p.Widgets[0].RowColors[0] != ColorRed {
		t.Fatalf("unexpected table widget %+v", p.Widgets[0])
	}
	if p.Widgets[1].Kind != WidgetKindTimeline || len(p.Widgets[1].Events) != 1 {
		t.Fatalf("unexpected timeline widget %+v", p.Widgets[1])
	}
}
