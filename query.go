package godebugbar

import (
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
)

// Query describes one executed database statement.
type Query struct {
	// SQL is the statement as executed, with values inlined when known.
	SQL string
	// Template is the statement with placeholders. It groups repeats for N+1
	// detection and defaults to SQL.
	Template string
	// Args are the bound values.
	Args []any
	// RowsAffected is the number of rows returned or changed, when known.
	RowsAffected int64
	// Duration is how long the statement took.
	Duration time.Duration
	// StartedAt is when the statement started; it defaults to now minus Duration.
	StartedAt time.Time
	// Err is the error the statement returned.
	Err error
	// Source is the caller as "file:line"; it is found when empty.
	Source string
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
