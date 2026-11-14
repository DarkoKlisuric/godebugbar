package collector

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const requestCollectorName = "request"

type RequestCollector struct{}

func NewRequestCollector() *RequestCollector {
	return &RequestCollector{}
}

func (c *RequestCollector) Name() string {
	return requestCollectorName
}

func (c *RequestCollector) Collect(_ context.Context, request *http.Request, response *contract.ResponseInfo) *contract.Section {
	route := response.Route
	if route == "" {
		route = request.URL.Path
	}

	durationMs := float64(response.Duration.Microseconds()) / 1000.0
	durationText := fmt.Sprintf("%.2f ms", durationMs)
	statusText := strconv.Itoa(response.Status)

	panel := contract.NewPanel("Request").
		AddMetrics(
			contract.NewMetric("Status", fmt.Sprintf("%d %s", response.Status, http.StatusText(response.Status))),
			contract.NewMetric("Method", request.Method),
			contract.NewMetric("Route", route),
			contract.NewMetric("Duration", durationText),
			contract.NewMetric("Response Size", fmt.Sprintf("%d B", response.Size)),
		).
		AddKeyValue("Request Headers", flattenHeaders(request.Header)).
		AddKeyValue("Response Headers", flattenHeaders(response.Headers))

	if len(request.URL.Query()) > 0 {
		queryPairs := make(map[string]string, len(request.URL.Query()))
		for key, values := range request.URL.Query() {
			queryPairs[key] = strings.Join(values, ", ")
		}

		panel.AddKeyValue("Query Parameters", queryPairs)
	}

	if len(response.Errors) > 0 {
		rows := make([][]string, 0, len(response.Errors))
		for i, message := range response.Errors {
			rows = append(rows, []string{strconv.Itoa(i + 1), message})
		}

		panel.AddTable("Errors", []string{"#", "Error"}, rows)
	}

	return contract.NewSection(
		contract.ToolbarItem{
			Label: fmt.Sprintf("%s %s", statusText, request.Method),
			Value: fmt.Sprintf("%s (%s)", route, durationText),
			Color: contract.StatusColor(response.Status),
			Tooltip: map[string]string{
				"Path":     request.URL.Path,
				"Route":    route,
				"Duration": durationText,
			},
		},
		panel,
	)
}

func flattenHeaders(header http.Header) map[string]string {
	pairs := make(map[string]string, len(header))
	for key, values := range header {
		pairs[key] = strings.Join(values, ", ")
	}

	return pairs
}
