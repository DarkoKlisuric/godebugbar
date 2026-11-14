package godebugbar

import "time"

type Profile struct {
	ID         string     `json:"id"`
	Method     string     `json:"method"`
	Path       string     `json:"path"`
	Route      string     `json:"route"`
	Status     int        `json:"status"`
	DurationMs float64    `json:"duration_ms"`
	Size       int        `json:"size"`
	StartedAt  time.Time  `json:"started_at"`
	Sections   []*Section `json:"sections"`
}
