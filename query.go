package godebugbar

import (
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
)

type Query struct {
	SQL          string
	Template     string
	Args         []any
	RowsAffected int64
	Duration     time.Duration
	StartedAt    time.Time
	Err          error
	Source       string
}

func (q *Query) entry() contextstate.QueryEntry {
	entry := contextstate.QueryEntry{
		SQL:          q.SQL,
		Template:     q.Template,
		Vars:         formatArgs(q.Args),
		RowsAffected: q.RowsAffected,
		Duration:     q.Duration,
		Source:       q.Source,
		StartedAt:    q.StartedAt,
	}
	if q.Err != nil {
		entry.Error = q.Err.Error()
	}

	return entry
}
