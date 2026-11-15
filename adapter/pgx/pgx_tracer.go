package pgxadapter

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

type queryStartKey struct{}

type queryStart struct {
	at   time.Time
	sql  string
	args []any
}

// PgxTracer is a pgx.QueryTracer that records queries run with a profiled
// request's context.
type PgxTracer struct{}

// NewTracer returns a PgxTracer.
func NewTracer() *PgxTracer {
	return &PgxTracer{}
}

// Attach registers the query collector.
func Attach() godebugbar.Option {
	return func(e *engine.Engine) {
		e.AddCollectors(collector.NewQueryCollector())
	}
}

// TraceQueryStart remembers when the query started.
func (t *PgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if contextstate.RecorderFromContext(ctx) == nil {
		return ctx
	}

	return context.WithValue(ctx, queryStartKey{}, &queryStart{at: time.Now(), sql: data.SQL, args: data.Args})
}

// TraceQueryEnd records the finished query.
func (t *PgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	rec := contextstate.RecorderFromContext(ctx)
	start, ok := ctx.Value(queryStartKey{}).(*queryStart)
	if rec == nil || !ok {
		return
	}

	godebugbar.FromContext(ctx).AddQuery(godebugbar.Query{
		SQL:          start.sql,
		Template:     start.sql,
		Args:         start.args,
		RowsAffected: data.CommandTag.RowsAffected(),
		Duration:     time.Since(start.at),
		StartedAt:    start.at,
		Err:          data.Err,
	})
}
