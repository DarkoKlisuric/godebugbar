package collector

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const queryCollectorName = "database"

type QueryCollector struct{}

func NewQueryCollector() *QueryCollector {
	return &QueryCollector{}
}

func (c *QueryCollector) Name() string {
	return queryCollectorName
}

func (c *QueryCollector) Collect(ctx context.Context, _ *http.Request, _ *contract.ResponseInfo) *contract.Section {
	recorder := contextstate.RecorderFromContext(ctx)
	config := contextstate.NewRecorder().Config()
	var queries []contextstate.QueryEntry
	var dropped int
	if recorder != nil {
		queries = recorder.Queries()
		config = recorder.Config()
		dropped = recorder.Dropped().Queries
	}

	stats := contextstate.AnalyzeQueries(queries, config)
	rows, colors := queryRows(queries, stats)

	totalMs := formatMs(stats.Total)
	panel := contract.NewPanel("Database Queries").
		AddMetrics(
			contract.NewMetric("Total Queries", strconv.Itoa(len(queries)+dropped)),
			contract.NewMetric("Total Duration", totalMs),
			contract.NewMetric("Slow Queries", strconv.Itoa(stats.Slow)),
			contract.NewMetric("Duplicate Queries", strconv.Itoa(stats.Duplicates)),
			contract.NewMetric("N+1 Suspects", strconv.Itoa(stats.NPlusOne)),
			contract.NewMetric("Failed Queries", strconv.Itoa(stats.Failed)),
		).
		AddColoredTable(
			"Executed SQL Statements",
			[]string{"#", "SQL", "Variables", "Rows", "Duration", "Source", "Status"},
			rows,
			colors,
		)

	if repeated := repeatedTemplates(stats); len(repeated) > 0 {
		panel.AddTable("Repeated Statements", []string{"Count", "Statement"}, repeated)
	}
	if dropped > 0 {
		panel.AddKeyValue("Limit", map[string]string{
			"Not recorded": fmt.Sprintf("%d queries over the per-request limit of %d", dropped, config.EntryLimit),
		})
	}

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "DB",
			Value: fmt.Sprintf("%d (%s)", len(queries)+dropped, totalMs),
			Color: queryColor(stats, len(queries)),
			Tooltip: map[string]string{
				"Queries":      strconv.Itoa(len(queries) + dropped),
				"Slow":         strconv.Itoa(stats.Slow),
				"Duplicates":   strconv.Itoa(stats.Duplicates),
				"N+1 suspects": strconv.Itoa(stats.NPlusOne),
				"Duration":     totalMs,
			},
		},
		panel,
	)
}

func queryRows(queries []contextstate.QueryEntry, stats *contextstate.QueryAnalysis) ([][]string, []contract.BadgeColor) {
	rows := make([][]string, 0, len(queries))
	colors := make([]contract.BadgeColor, 0, len(queries))
	for i, q := range queries {
		status, color := queryStatus(q, stats)
		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			q.SQL,
			strings.Join(q.Vars, ", "),
			strconv.FormatInt(q.RowsAffected, 10),
			formatMs(q.Duration),
			q.Source,
			status,
		})
		colors = append(colors, color)
	}

	return rows, colors
}

func queryStatus(q contextstate.QueryEntry, stats *contextstate.QueryAnalysis) (string, contract.BadgeColor) {
	var labels []string
	color := contract.ColorDefault
	if q.Error != "" {
		return q.Error, contract.ColorRed
	}
	if stats.IsSlow(q) {
		labels = append(labels, "SLOW")
		color = contract.ColorYellow
	}
	if stats.IsDuplicate(q) {
		labels = append(labels, "DUPLICATE")
		color = contract.ColorYellow
	}
	if stats.IsNPlusOne(q) {
		labels = append(labels, fmt.Sprintf("N+1 (%d×)", stats.Repeats(q)))
		color = contract.ColorYellow
	}
	if len(labels) == 0 {
		return "OK", color
	}

	return strings.Join(labels, ", "), color
}

func repeatedTemplates(stats *contextstate.QueryAnalysis) [][]string {
	type repeated struct {
		template string
		count    int
	}
	var list []repeated
	for template, count := range stats.RepeatedTemplates() {
		list = append(list, repeated{template: template, count: count})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].template < list[j].template
	})

	rows := make([][]string, 0, len(list))
	for _, r := range list {
		rows = append(rows, []string{strconv.Itoa(r.count), r.template})
	}

	return rows
}

func queryColor(stats *contextstate.QueryAnalysis, count int) contract.BadgeColor {
	switch {
	case stats.Failed > 0:
		return contract.ColorRed
	case stats.Slow > 0 || stats.Duplicates > 0 || stats.NPlusOne > 0:
		return contract.ColorYellow
	case count > 0:
		return contract.ColorGreen
	default:
		return contract.ColorDefault
	}
}
