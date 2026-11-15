package gormadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

const (
	pluginName           = "godebugbar"
	startTimeInstanceKey = "godebugbar:start_time"
	explainPrefix        = "EXPLAIN "
)

type gormPlugin struct{}

// Attach registers the query collector, installs Plugin on db and enables
// EXPLAIN for recorded SELECT statements. It panics when the plugin cannot
// be registered.
func Attach(db *gorm.DB) godebugbar.Option {
	return func(e *engine.Engine) {
		if db == nil {
			return
		}
		e.AddCollectors(collector.NewQueryCollector())
		e.SetExplainer(explainer(db))
		if err := db.Use(Plugin()); err != nil && !errors.Is(err, gorm.ErrRegistered) {
			panic(fmt.Sprintf("godebugbar: registering the gorm plugin: %v", err))
		}
	}
}

// Plugin returns the GORM plugin that records queries, for a *gorm.DB built
// apart from the Profiler. Pair it with godebugbar.WithQueryCollector.
func Plugin() gorm.Plugin {
	return &gormPlugin{}
}

func (p *gormPlugin) Name() string {
	return pluginName
}

func (p *gormPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	steps := []error{
		cb.Create().Before("*").Register("godebugbar:before_create", p.before),
		cb.Create().After("*").Register("godebugbar:after_create", p.after),
		cb.Query().Before("*").Register("godebugbar:before_query", p.before),
		cb.Query().After("*").Register("godebugbar:after_query", p.after),
		cb.Update().Before("*").Register("godebugbar:before_update", p.before),
		cb.Update().After("*").Register("godebugbar:after_update", p.after),
		cb.Delete().Before("*").Register("godebugbar:before_delete", p.before),
		cb.Delete().After("*").Register("godebugbar:after_delete", p.after),
		cb.Row().Before("*").Register("godebugbar:before_row", p.before),
		cb.Row().After("*").Register("godebugbar:after_row", p.after),
		cb.Raw().Before("*").Register("godebugbar:before_raw", p.before),
		cb.Raw().After("*").Register("godebugbar:after_raw", p.after),
	}

	return errors.Join(steps...)
}

func (p *gormPlugin) before(db *gorm.DB) {
	if db == nil || db.Statement == nil {
		return
	}
	db.InstanceSet(startTimeInstanceKey, time.Now())
}

func (p *gormPlugin) after(db *gorm.DB) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil {
		return
	}
	rec := contextstate.RecorderFromContext(db.Statement.Context)
	if rec == nil {
		return
	}

	var startedAt time.Time
	var duration time.Duration
	if v, ok := db.InstanceGet(startTimeInstanceKey); ok {
		if start, ok := v.(time.Time); ok {
			startedAt = start
			duration = time.Since(start)
		}
	}

	template := db.Statement.SQL.String()
	sqlText := template
	if !rec.Config().HideSQLValues {
		if explained := db.Explain(template, db.Statement.Vars...); explained != "" {
			sqlText = explained
		}
	}

	vars := make([]string, len(db.Statement.Vars))
	for i, v := range db.Statement.Vars {
		vars[i] = fmt.Sprint(v)
	}

	errText := ""
	if db.Error != nil {
		errText = db.Error.Error()
	}

	rec.RecordQuery(contextstate.QueryEntry{
		SQL:          sqlText,
		Template:     template,
		Vars:         vars,
		RowsAffected: db.Statement.RowsAffected,
		Duration:     duration,
		Error:        errText,
		StartedAt:    startedAt,
	})
}

func explainer(db *gorm.DB) engine.Explainer {
	return func(ctx context.Context, statement string) ([]string, error) {
		var plan []string
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			rows, err := tx.Raw(explainPrefix + statement).Rows()
			if err != nil {
				return err
			}
			defer func() { _ = rows.Close() }()

			columns, err := rows.Columns()
			if err != nil {
				return err
			}
			for rows.Next() {
				values := make([]sql.NullString, len(columns))
				targets := make([]any, len(columns))
				for i := range values {
					targets[i] = &values[i]
				}
				if err := rows.Scan(targets...); err != nil {
					return err
				}
				for _, v := range values {
					if v.Valid {
						plan = append(plan, v.String)
					}
				}
			}

			return rows.Err()
		}, &sql.TxOptions{ReadOnly: true})

		return plan, err
	}
}
