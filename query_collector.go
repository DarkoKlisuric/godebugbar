package godebugbar

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const queryCollectorName = "database"

type QueryCollector struct{}

func NewQueryCollector() *QueryCollector {
	return &QueryCollector{}
}

func (c *QueryCollector) Name() string {
	return queryCollectorName
}

func (c *QueryCollector) Collect(ctx context.Context, _ *http.Request, _ *ResponseInfo) *Section {
	var queries []QueryEntry
	if recorder := RecorderFromContext(ctx); recorder != nil {
		queries = recorder.Queries()
	}

	var totalDuration time.Duration
	countsBySQL := make(map[string]int, len(queries))
	failedCount := 0

	for _, query := range queries {
		totalDuration += query.Duration
		countsBySQL[query.SQL]++
		if query.Error != "" {
			failedCount++
		}
	}

	duplicateCount := 0
	for _, count := range countsBySQL {
		if count > 1 {
			duplicateCount += count - 1
		}
	}

	rows := make([][]string, 0, len(queries))
	for i, query := range queries {
		status := "OK"
		if query.Error != "" {
			status = query.Error
		} else if countsBySQL[query.SQL] > 1 {
			status = "DUPLICATE"
		}

		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			query.SQL,
			strings.Join(query.Vars, ", "),
			strconv.FormatInt(query.RowsAffected, 10),
			fmt.Sprintf("%.2f ms", float64(query.Duration.Microseconds())/1000.0),
			query.Source,
			status,
		})
	}

	totalMs := fmt.Sprintf("%.2f ms", float64(totalDuration.Microseconds())/1000.0)

	panel := NewPanel("Database Queries").
		AddMetrics(
			NewMetric("Total Queries", strconv.Itoa(len(queries))),
			NewMetric("Total Duration", totalMs),
			NewMetric("Duplicate Queries", strconv.Itoa(duplicateCount)),
			NewMetric("Failed Queries", strconv.Itoa(failedCount)),
		).
		AddTable(
			"Executed SQL Statements",
			[]string{"#", "SQL", "Variables", "Rows", "Duration", "Source", "Status"},
			rows,
		)

	color := ColorDefault
	switch {
	case failedCount > 0:
		color = ColorRed
	case duplicateCount > 0:
		color = ColorYellow
	case len(queries) > 0:
		color = ColorGreen
	}

	return NewSection(
		ToolbarItem{
			Label: "DB",
			Value: fmt.Sprintf("%d (%s)", len(queries), totalMs),
			Color: color,
			Tooltip: map[string]string{
				"Queries":    strconv.Itoa(len(queries)),
				"Duplicates": strconv.Itoa(duplicateCount),
				"Duration":   totalMs,
			},
		},
		panel,
	)
}
