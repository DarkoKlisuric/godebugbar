package contextstate

import "time"

type SpanEntry struct {
	Name      string        `json:"name"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration"`
}
