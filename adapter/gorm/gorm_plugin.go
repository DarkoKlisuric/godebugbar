package gormadapter

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

const startTimeInstanceKey = "godebugbar:start_time"

type gormPlugin struct{}

func Attach(db *gorm.DB) godebugbar.Option {
	return func(p *godebugbar.Profiler) {
		if db == nil || p == nil {
			return
		}
		if eng := engine.Lookup(p); eng != nil {
			eng.AddCollectors(collector.NewQueryCollector())
		}
		_ = db.Use(&gormPlugin{})
	}
}

func (p *gormPlugin) Name() string {
	return "godebugbar"
}

func (p *gormPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()

	if err := cb.Create().Before("*").Register("godebugbar:before_create", p.before); err != nil {
		return err
	}
	if err := cb.Create().After("*").Register("godebugbar:after_create", p.after); err != nil {
		return err
	}
	if err := cb.Query().Before("*").Register("godebugbar:before_query", p.before); err != nil {
		return err
	}
	if err := cb.Query().After("*").Register("godebugbar:after_query", p.after); err != nil {
		return err
	}
	if err := cb.Update().Before("*").Register("godebugbar:before_update", p.before); err != nil {
		return err
	}
	if err := cb.Update().After("*").Register("godebugbar:after_update", p.after); err != nil {
		return err
	}
	if err := cb.Delete().Before("*").Register("godebugbar:before_delete", p.before); err != nil {
		return err
	}
	if err := cb.Delete().After("*").Register("godebugbar:after_delete", p.after); err != nil {
		return err
	}
	if err := cb.Row().Before("*").Register("godebugbar:before_row", p.before); err != nil {
		return err
	}
	if err := cb.Row().After("*").Register("godebugbar:after_row", p.after); err != nil {
		return err
	}
	if err := cb.Raw().Before("*").Register("godebugbar:before_raw", p.before); err != nil {
		return err
	}
	return cb.Raw().After("*").Register("godebugbar:after_raw", p.after)
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

	sql := db.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	if sql == "" {
		sql = db.Statement.SQL.String()
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
		SQL:          sql,
		Vars:         vars,
		RowsAffected: db.Statement.RowsAffected,
		Duration:     duration,
		Source:       utils.FileWithLineNum(),
		Error:        errText,
		StartedAt:    startedAt,
	})
}
