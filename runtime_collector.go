package godebugbar

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
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

func (c *RuntimeCollector) Collect(_ context.Context, _ *http.Request, _ *ResponseInfo) *Section {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	goroutines := runtime.NumGoroutine()
	heapMB := float64(stats.HeapAlloc) / bytesPerMegabyte
	sysMB := float64(stats.Sys) / bytesPerMegabyte

	heapText := fmt.Sprintf("%.2f MB", heapMB)
	sysText := fmt.Sprintf("%.2f MB", sysMB)

	panel := NewPanel("Go Runtime").
		AddMetrics(
			NewMetric("Heap Alloc", heapText),
			NewMetric("System Memory", sysText),
			NewMetric("Goroutines", strconv.Itoa(goroutines)),
			NewMetric("GC Cycles", strconv.FormatUint(uint64(stats.NumGC), 10)),
		).
		AddKeyValue("Environment", map[string]string{
			"Go Version": runtime.Version(),
			"OS / Arch":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"CPUs":       strconv.Itoa(runtime.NumCPU()),
		})

	return NewSection(
		ToolbarItem{
			Label: "Runtime",
			Value: fmt.Sprintf("%s | %d gr", heapText, goroutines),
			Color: ColorDefault,
			Tooltip: map[string]string{
				"Heap":       heapText,
				"Goroutines": strconv.Itoa(goroutines),
				"Go":         runtime.Version(),
			},
		},
		panel,
	)
}
