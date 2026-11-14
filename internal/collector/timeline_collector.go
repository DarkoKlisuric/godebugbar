package collector

import (
	"context"
	"net/http"
	"sort"
	"strconv"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const (
	timelineCollectorName = "timeline"
	categoryRequest       = "request"
	categorySpan          = "span"
	categoryQuery         = "query"
	categoryHTTP          = "http"
	timelineLabelLimit    = 120
)

type TimelineCollector struct{}

func NewTimelineCollector() *TimelineCollector {
	return &TimelineCollector{}
}

func (c *TimelineCollector) Name() string {
	return timelineCollectorName
}

func (c *TimelineCollector) Collect(ctx context.Context, request *http.Request, response *contract.ResponseInfo) *contract.Section {
	recorder := contextstate.RecorderFromContext(ctx)
	if recorder == nil {
		return nil
	}
	start := recorder.Start()
	events := make([]contract.TimelineEvent, 0)

	for _, span := range recorder.Spans() {
		events = append(events, contract.TimelineEvent{
			Label: span.Name, Category: categorySpan,
			StartMs: offsetMs(start, span.StartedAt), DurationMs: msFloat(span.Duration),
		})
	}
	config := recorder.Config()
	for _, q := range recorder.Queries() {
		color := contract.ColorDefault
		if q.Error != "" {
			color = contract.ColorRed
		} else if q.Duration >= config.SlowQuery {
			color = contract.ColorYellow
		}
		events = append(events, contract.TimelineEvent{
			Label: shorten(templateOf(q)), Category: categoryQuery,
			StartMs: offsetMs(start, q.StartedAt), DurationMs: msFloat(q.Duration), Color: color,
		})
	}
	for _, call := range recorder.HTTPCalls() {
		events = append(events, contract.TimelineEvent{
			Label: call.Method + " " + shorten(call.URL), Category: categoryHTTP,
			StartMs: offsetMs(start, call.StartedAt), DurationMs: msFloat(call.Duration),
		})
	}
	if len(events) == 0 {
		return nil
	}

	sort.SliceStable(events, func(i, j int) bool { return events[i].StartMs < events[j].StartMs })
	events = append([]contract.TimelineEvent{{
		Label: request.Method + " " + request.URL.Path, Category: categoryRequest,
		StartMs: 0, DurationMs: msFloat(response.Duration), Color: contract.StatusColor(response.Status),
	}}, events...)

	totalText := formatMs(response.Duration)
	panel := contract.NewPanel("Timeline").AddTimeline("Request Timeline", events)

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "Timeline",
			Value: strconv.Itoa(len(events)-1) + " events",
			Color: contract.ColorDefault,
			Tooltip: map[string]string{
				"Events":   strconv.Itoa(len(events) - 1),
				"Duration": totalText,
			},
		},
		panel,
	)
}

func shorten(s string) string {
	runes := []rune(s)
	if len(runes) <= timelineLabelLimit {
		return s
	}

	return string(runes[:timelineLabelLimit]) + "…"
}
