package godebugbar

import "github.com/DarkoKlisuric/godebugbar/internal/contract"

// Panel building blocks.
type (
	// Panel is an inspector tab made of widgets, built with the Add* methods.
	Panel = contract.Panel
	// Widget is one block of a Panel.
	Widget = contract.Widget
	// WidgetKind says how a Widget is drawn.
	WidgetKind = contract.WidgetKind
	// Metric is one labelled value in a metrics widget.
	Metric = contract.Metric
	// TimelineEvent is one bar in a timeline widget.
	TimelineEvent = contract.TimelineEvent
)

// Widget kinds.
const (
	WidgetKindMetrics  = contract.WidgetKindMetrics
	WidgetKindKeyValue = contract.WidgetKindKeyValue
	WidgetKindTable    = contract.WidgetKindTable
	WidgetKindCode     = contract.WidgetKindCode
	WidgetKindTimeline = contract.WidgetKindTimeline
)

// NewPanel creates an empty Panel.
func NewPanel(title string) *Panel {
	return contract.NewPanel(title)
}

// NewMetric creates a Metric.
func NewMetric(label string, value string) Metric {
	return contract.NewMetric(label, value)
}
