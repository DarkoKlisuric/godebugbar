package contextstate

import "time"

type QueryEntry struct {
	SQL          string        `json:"sql"`
	Template     string        `json:"template,omitempty"`
	Vars         []string      `json:"vars,omitempty"`
	RowsAffected int64         `json:"rows_affected"`
	Duration     time.Duration `json:"duration"`
	Source       string        `json:"source,omitempty"`
	Error        string        `json:"error,omitempty"`
	StartedAt    time.Time     `json:"started_at"`
}
