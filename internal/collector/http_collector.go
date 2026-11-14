package collector

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const httpCollectorName = "http_client"

type HTTPCollector struct{}

func NewHTTPCollector() *HTTPCollector {
	return &HTTPCollector{}
}

func (c *HTTPCollector) Name() string {
	return httpCollectorName
}

func (c *HTTPCollector) Collect(ctx context.Context, _ *http.Request, _ *contract.ResponseInfo) *contract.Section {
	recorder := contextstate.RecorderFromContext(ctx)
	if recorder == nil {
		return nil
	}
	calls := recorder.HTTPCalls()
	if len(calls) == 0 {
		return nil
	}

	var total time.Duration
	failed := 0
	rows := make([][]string, 0, len(calls))
	colors := make([]contract.BadgeColor, 0, len(calls))
	for i, call := range calls {
		total += call.Duration
		color := contract.StatusColor(call.Status)
		status := strconv.Itoa(call.Status)
		if call.Error != "" {
			color = contract.ColorRed
			status = call.Error
		}
		if color == contract.ColorRed || color == contract.ColorYellow {
			failed++
		}
		rows = append(rows, []string{strconv.Itoa(i + 1), call.Method, call.URL, status, formatMs(call.Duration)})
		colors = append(colors, color)
	}

	color := contract.ColorGreen
	if failed > 0 {
		color = contract.ColorYellow
	}
	totalText := formatMs(total)

	panel := contract.NewPanel("Outgoing HTTP").
		AddMetrics(
			contract.NewMetric("Calls", strconv.Itoa(len(calls))),
			contract.NewMetric("Total Duration", totalText),
			contract.NewMetric("Failed", strconv.Itoa(failed)),
		).
		AddColoredTable("Calls", []string{"#", "Method", "URL", "Status", "Duration"}, rows, colors)

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "HTTP",
			Value: fmt.Sprintf("%d (%s)", len(calls), totalText),
			Color: color,
			Tooltip: map[string]string{
				"Calls":    strconv.Itoa(len(calls)),
				"Failed":   strconv.Itoa(failed),
				"Duration": totalText,
			},
		},
		panel,
	)
}
