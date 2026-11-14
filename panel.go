package godebugbar

import "github.com/DarkoKlisuric/godebugbar/internal/contract"

type (
	Panel         = contract.Panel
	Widget        = contract.Widget
	WidgetKind    = contract.WidgetKind
	Metric        = contract.Metric
	TimelineEvent = contract.TimelineEvent
)

const (
	WidgetKindMetrics  = contract.WidgetKindMetrics
	WidgetKindKeyValue = contract.WidgetKindKeyValue
	WidgetKindTable    = contract.WidgetKindTable
	WidgetKindCode     = contract.WidgetKindCode
	WidgetKindTimeline = contract.WidgetKindTimeline
)

func NewPanel(title string) *Panel {
	return contract.NewPanel(title)
}

func NewMetric(label string, value string) Metric {
	return contract.NewMetric(label, value)
}
