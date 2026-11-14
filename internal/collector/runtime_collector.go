package collector

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

const (
	runtimeCollectorName = "runtime"
	bytesPerMegabyte     = 1024 * 1024
)

type RuntimeCollector struct{}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{}
}

func (c *RuntimeCollector) Name() string {
	return runtimeCollectorName
}

func (c *RuntimeCollector) Collect(_ context.Context, _ *http.Request, _ *contract.ResponseInfo) *contract.Section {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	goroutines := runtime.NumGoroutine()
	heapMB := float64(stats.HeapAlloc) / bytesPerMegabyte
	sysMB := float64(stats.Sys) / bytesPerMegabyte

	heapText := fmt.Sprintf("%.2f MB", heapMB)
	sysText := fmt.Sprintf("%.2f MB", sysMB)

	panel := contract.NewPanel("Go Runtime").
		AddMetrics(
			contract.NewMetric("Heap Alloc", heapText),
			contract.NewMetric("System Memory", sysText),
			contract.NewMetric("Goroutines", strconv.Itoa(goroutines)),
			contract.NewMetric("GC Cycles", strconv.FormatUint(uint64(stats.NumGC), 10)),
		).
		AddKeyValue("Environment", map[string]string{
			"Go Version": runtime.Version(),
			"OS / Arch":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"CPUs":       strconv.Itoa(runtime.NumCPU()),
		})

	return contract.NewSection(
		contract.ToolbarItem{
			Label: "Runtime",
			Value: fmt.Sprintf("%s | %d gr", heapText, goroutines),
			Color: contract.ColorDefault,
			Tooltip: map[string]string{
				"Heap":       heapText,
				"Goroutines": strconv.Itoa(goroutines),
				"Go":         runtime.Version(),
			},
		},
		panel,
	)
}
