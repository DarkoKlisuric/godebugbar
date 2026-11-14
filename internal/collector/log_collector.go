package collector

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const logCollectorName = "logs"

type LogCollector struct{}

func NewLogCollector() *LogCollector {
	return &LogCollector{}
}

func (c *LogCollector) Name() string {
	return logCollectorName
}

func (c *LogCollector) Collect(ctx context.Context, _ *http.Request, _ *contract.ResponseInfo) *contract.Section {
	recorder := contextstate.RecorderFromContext(ctx)
	if recorder == nil {
		return nil
	}
	logs := recorder.Logs()
	if len(logs) == 0 {
		return nil
	}

	rows := make([][]string, 0, len(logs))
	colors := make([]contract.BadgeColor, 0, len(logs))
	color := contract.ColorDefault
	warnings := 0
	for _, entry := range logs {
		rowColor := levelColor(entry.Level)
		if rowColor != contract.ColorDefault {
			warnings++
		}
		color = worse(color, rowColor)
		rows = append(rows, []string{
			strconv.FormatFloat(offsetMs(recorder.Start(), entry.Time), 'f', 2, 64) + " ms",
			entry.Level,
			entry.Message,
			formatAttrs(entry.Attrs),
		})
		colors = append(colors, rowColor)
	}

	panel := contract.NewPanel("Logs").
		AddMetrics(
			contract.NewMetric("Messages", strconv.Itoa(len(logs))),
			contract.NewMetric("Warnings And Errors", strconv.Itoa(warnings)),
		).
		AddColoredTable("Log Messages", []string{"At", "Level", "Message", "Attributes"}, rows, colors)

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "Logs",
			Value: strconv.Itoa(len(logs)),
			Color: color,
			Tooltip: map[string]string{
				"Messages":            strconv.Itoa(len(logs)),
				"Warnings and errors": strconv.Itoa(warnings),
			},
		},
		panel,
	)
}

func levelColor(level string) contract.BadgeColor {
	switch {
	case strings.HasPrefix(level, "ERROR"):
		return contract.ColorRed
	case strings.HasPrefix(level, "WARN"):
		return contract.ColorYellow
	default:
		return contract.ColorDefault
	}
}

func worse(a, b contract.BadgeColor) contract.BadgeColor {
	rank := map[contract.BadgeColor]int{contract.ColorDefault: 0, contract.ColorGreen: 1, contract.ColorYellow: 2, contract.ColorRed: 3}
	if rank[b] > rank[a] {
		return b
	}

	return a
}

func formatAttrs(attrs map[string]string) string {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+attrs[key])
	}

	return strings.Join(parts, " ")
}
