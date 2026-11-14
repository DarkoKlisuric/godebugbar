package collector

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"runtime/metrics"
	"strconv"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const (
	runtimeCollectorName = "runtime"
	bytesPerMegabyte     = 1024 * 1024

	metricHeapObjects = "/memory/classes/heap/objects:bytes"
	metricTotalMemory = "/memory/classes/total:bytes"
	metricGoroutines  = "/sched/goroutines:goroutines"
	metricGCCycles    = "/gc/cycles/total:gc-cycles"
)

type RuntimeCollector struct{}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{}
}

func (c *RuntimeCollector) Name() string {
	return runtimeCollectorName
}

func (c *RuntimeCollector) Collect(_ context.Context, _ *http.Request, _ *contract.ResponseInfo) *contract.Section {
	samples := []metrics.Sample{
		{Name: metricHeapObjects},
		{Name: metricTotalMemory},
		{Name: metricGoroutines},
		{Name: metricGCCycles},
	}
	metrics.Read(samples)

	heapText := fmt.Sprintf("%.2f MB", float64(sampleUint(samples[0]))/bytesPerMegabyte)
	sysText := fmt.Sprintf("%.2f MB", float64(sampleUint(samples[1]))/bytesPerMegabyte)
	goroutineText := strconv.FormatUint(sampleUint(samples[2]), 10)
	if goroutineText == "0" {
		goroutineText = strconv.Itoa(runtime.NumGoroutine())
	}

	panel := contract.NewPanel("Go Runtime").
		AddMetrics(
			contract.NewMetric("Heap In Use", heapText),
			contract.NewMetric("Total Memory", sysText),
			contract.NewMetric("Goroutines", goroutineText),
			contract.NewMetric("GC Cycles", strconv.FormatUint(sampleUint(samples[3]), 10)),
		).
		AddKeyValue("Environment", map[string]string{
			"Go Version": runtime.Version(),
			"OS / Arch":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"CPUs":       strconv.Itoa(runtime.NumCPU()),
		})

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "Runtime",
			Value: fmt.Sprintf("%s | %s gr", heapText, goroutineText),
			Color: contract.ColorDefault,
			Tooltip: map[string]string{
				"Heap":       heapText,
				"Goroutines": goroutineText,
				"Go":         runtime.Version(),
			},
		},
		panel,
	)
}

func sampleUint(s metrics.Sample) uint64 {
	if s.Value.Kind() != metrics.KindUint64 {
		return 0
	}

	return s.Value.Uint64()
}
