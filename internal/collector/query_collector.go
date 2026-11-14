package collector

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const queryCollectorName = "database"

type QueryCollector struct{}

type queryStats struct {
	total      time.Duration
	failed     int
	slow       int
	duplicates int
	nPlusOne   int
	byTemplate map[string]int
	bySQL      map[string]int
}

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

	stats := analyzeQueries(queries, config)
	rows, colors := queryRows(queries, stats, config)

	totalMs := formatMs(stats.total)
	panel := contract.NewPanel("Database Queries").
		AddMetrics(
			contract.NewMetric("Total Queries", strconv.Itoa(len(queries)+dropped)),
			contract.NewMetric("Total Duration", totalMs),
			contract.NewMetric("Slow Queries", strconv.Itoa(stats.slow)),
			contract.NewMetric("Duplicate Queries", strconv.Itoa(stats.duplicates)),
			contract.NewMetric("N+1 Suspects", strconv.Itoa(stats.nPlusOne)),
			contract.NewMetric("Failed Queries", strconv.Itoa(stats.failed)),
		).
		AddColoredTable(
			"Executed SQL Statements",
			[]string{"#", "SQL", "Variables", "Rows", "Duration", "Source", "Status"},
			rows,
			colors,
		)

	if repeated := repeatedTemplates(stats, config); len(repeated) > 0 {
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
				"Slow":         strconv.Itoa(stats.slow),
				"Duplicates":   strconv.Itoa(stats.duplicates),
				"N+1 suspects": strconv.Itoa(stats.nPlusOne),
				"Duration":     totalMs,
			},
		},
		panel,
	)
}

func analyzeQueries(queries []contextstate.QueryEntry, config contextstate.Config) queryStats {
	stats := queryStats{
		byTemplate: make(map[string]int, len(queries)),
		bySQL:      make(map[string]int, len(queries)),
	}
	for _, q := range queries {
		stats.total += q.Duration
		stats.byTemplate[templateOf(q)]++
		stats.bySQL[q.SQL+"\x00"+strings.Join(q.Vars, "\x00")]++
		if q.Error != "" {
			stats.failed++
		}
		if q.Duration >= config.SlowQuery {
			stats.slow++
		}
	}
	for _, count := range stats.bySQL {
		if count > 1 {
			stats.duplicates += count - 1
		}
	}
	for _, count := range stats.byTemplate {
		if count >= config.NPlusOneThreshold {
			stats.nPlusOne++
		}
	}

	return stats
}

func queryRows(queries []contextstate.QueryEntry, stats queryStats, config contextstate.Config) ([][]string, []contract.BadgeColor) {
	rows := make([][]string, 0, len(queries))
	colors := make([]contract.BadgeColor, 0, len(queries))
	for i, q := range queries {
		status, color := queryStatus(q, stats, config)
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

func queryStatus(q contextstate.QueryEntry, stats queryStats, config contextstate.Config) (string, contract.BadgeColor) {
	var labels []string
	color := contract.ColorDefault
	if q.Error != "" {
		return q.Error, contract.ColorRed
	}
	if q.Duration >= config.SlowQuery {
		labels = append(labels, "SLOW")
		color = contract.ColorYellow
	}
	if stats.bySQL[q.SQL+"\x00"+strings.Join(q.Vars, "\x00")] > 1 {
		labels = append(labels, "DUPLICATE")
		color = contract.ColorYellow
	}
	if count := stats.byTemplate[templateOf(q)]; count >= config.NPlusOneThreshold {
		labels = append(labels, fmt.Sprintf("N+1 (%d×)", count))
		color = contract.ColorYellow
	}
	if len(labels) == 0 {
		return "OK", color
	}

	return strings.Join(labels, ", "), color
}

func repeatedTemplates(stats queryStats, config contextstate.Config) [][]string {
	type repeated struct {
		template string
		count    int
	}
	var list []repeated
	for template, count := range stats.byTemplate {
		if count >= config.NPlusOneThreshold {
			list = append(list, repeated{template: template, count: count})
		}
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

func queryColor(stats queryStats, count int) contract.BadgeColor {
	switch {
	case stats.failed > 0:
		return contract.ColorRed
	case stats.slow > 0 || stats.duplicates > 0 || stats.nPlusOne > 0:
		return contract.ColorYellow
	case count > 0:
		return contract.ColorGreen
	default:
		return contract.ColorDefault
	}
}

func templateOf(q contextstate.QueryEntry) string {
	if q.Template != "" {
		return q.Template
	}

	return q.SQL
}
