package contextstate

import "time"

type HTTPCallEntry struct {
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	Status    int           `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	StartedAt time.Time     `json:"started_at"`
}
