package gormadapter

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
)

const (
	pluginName      = "godebugbar:gorm"
	beforeCallback  = "godebugbar:before"
	afterCallback   = "godebugbar:after"
	startTimeKey    = "godebugbar:start_time"
	maxCallerFrames = 15
)

type GormPlugin struct{}

func NewGormPlugin() *GormPlugin {
	return &GormPlugin{}
}

func Attach(db *gorm.DB) godebugbar.Option {
	return func(_ *godebugbar.Profiler) {
		if db == nil {
			return
		}

		_ = db.Use(NewGormPlugin())
	}
}

func (p *GormPlugin) Name() string {
	return pluginName
}

func (p *GormPlugin) Initialize(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	callbacks := db.Callback()
	steps := []struct {
		before func(string, func(*gorm.DB)) error
		after  func(string, func(*gorm.DB)) error
	}{
		{before: callbacks.Create().Before("*").Register, after: callbacks.Create().After("*").Register},
		{before: callbacks.Query().Before("*").Register, after: callbacks.Query().After("*").Register},
		{before: callbacks.Update().Before("*").Register, after: callbacks.Update().After("*").Register},
		{before: callbacks.Delete().Before("*").Register, after: callbacks.Delete().After("*").Register},
		{before: callbacks.Row().Before("*").Register, after: callbacks.Row().After("*").Register},
		{before: callbacks.Raw().Before("*").Register, after: callbacks.Raw().After("*").Register},
	}

	for _, step := range steps {
		if err := step.before(beforeCallback, p.before); err != nil {
			return err
		}

		if err := step.after(afterCallback, p.after); err != nil {
			return err
		}
	}

	return nil
}

func (p *GormPlugin) before(db *gorm.DB) {
	if db == nil {
		return
	}

	db.InstanceSet(startTimeKey, time.Now().UTC())
}

func (p *GormPlugin) after(db *gorm.DB) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil {
		return
	}

	recorder := godebugbar.RecorderFromContext(db.Statement.Context)
	if recorder == nil {
		return
	}

	sqlText := strings.TrimSpace(db.Statement.SQL.String())
	if sqlText == "" {
		return
	}

	var startedAt time.Time
	var duration time.Duration
	if val, ok := db.InstanceGet(startTimeKey); ok {
		if start, valid := val.(time.Time); valid {
			startedAt = start
			duration = time.Since(start)
		}
	}

	vars := make([]string, 0, len(db.Statement.Vars))
	for _, variable := range db.Statement.Vars {
		vars = append(vars, fmt.Sprint(variable))
	}

	var errText string
	if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
		errText = db.Error.Error()
	}

	recorder.RecordQuery(godebugbar.QueryEntry{
		SQL:          sqlText,
		Vars:         vars,
		RowsAffected: db.Statement.RowsAffected,
		Duration:     duration,
		Source:       callerSource(),
		Error:        errText,
		StartedAt:    startedAt,
	})
}

func callerSource() string {
	var pcs [maxCallerFrames]uintptr
	count := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:count])

	for {
		frame, more := frames.Next()
		if frame.File != "" &&
			!strings.Contains(frame.File, "gorm.io/") &&
			!strings.Contains(frame.File, "godebugbar/adapter/gorm") {
			return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
		}

		if !more {
			break
		}
	}

	return ""
}
