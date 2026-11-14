package godebugbar

import "maps"

type WidgetKind string

const (
	WidgetKindMetrics  WidgetKind = "metrics"
	WidgetKindKeyValue WidgetKind = "key_value"
	WidgetKindTable    WidgetKind = "table"
	WidgetKindCode     WidgetKind = "code"
)

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func NewMetric(label, value string) Metric {
	return Metric{Label: label, Value: value}
}

type Widget struct {
	Kind    WidgetKind        `json:"kind"`
	Title   string            `json:"title,omitempty"`
	Metrics []Metric          `json:"metrics,omitempty"`
	Pairs   map[string]string `json:"pairs,omitempty"`
	Columns []string          `json:"columns,omitempty"`
	Rows    [][]string        `json:"rows,omitempty"`
	Code    string            `json:"code,omitempty"`
}

type Panel struct {
	Title   string   `json:"title"`
	Widgets []Widget `json:"widgets"`
}

func NewPanel(title string) *Panel {
	return &Panel{
		Title:   title,
		Widgets: make([]Widget, 0),
	}
}

func (p *Panel) AddMetrics(metrics ...Metric) *Panel {
	p.Widgets = append(p.Widgets, Widget{
		Kind:    WidgetKindMetrics,
		Metrics: metrics,
	})

	return p
}

func (p *Panel) AddKeyValue(title string, pairs map[string]string) *Panel {
	copied := make(map[string]string, len(pairs))
	maps.Copy(copied, pairs)

	p.Widgets = append(p.Widgets, Widget{
		Kind:  WidgetKindKeyValue,
		Title: title,
		Pairs: copied,
	})

	return p
}

func (p *Panel) AddTable(title string, columns []string, rows [][]string) *Panel {
	p.Widgets = append(p.Widgets, Widget{
		Kind:    WidgetKindTable,
		Title:   title,
		Columns: columns,
		Rows:    rows,
	})

	return p
}

func (p *Panel) AddCode(title, code string) *Panel {
	p.Widgets = append(p.Widgets, Widget{
		Kind:  WidgetKindCode,
		Title: title,
		Code:  code,
	})

	return p
}
