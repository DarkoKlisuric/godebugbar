package collector

import (
	"fmt"
	"time"
)

func formatMs(d time.Duration) string {
	return fmt.Sprintf("%.2f ms", msFloat(d))
}

func msFloat(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}

func offsetMs(start, at time.Time) float64 {
	if at.IsZero() || start.IsZero() {
		return 0
	}

	return msFloat(at.Sub(start))
}
