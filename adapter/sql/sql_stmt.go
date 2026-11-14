package sqladapter

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"
)

var errNamedArgs = errors.New("godebugbar: the wrapped driver does not support named arguments")

type sqlStmt struct {
	parent driver.Stmt
	query  string
}

func (s *sqlStmt) Close() error {
	return s.parent.Close()
}

func (s *sqlStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s *sqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.parent.Exec(args)
}

func (s *sqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.parent.Query(args)
}

func (s *sqlStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	var (
		res driver.Result
		err error
	)
	if ec, ok := s.parent.(driver.StmtExecContext); ok {
		res, err = ec.ExecContext(ctx, args)
	} else {
		var values []driver.Value
		if values, err = positional(args); err == nil {
			res, err = s.parent.Exec(values)
		}
	}
	record(ctx, s.query, args, start, rowsAffected(res), err)

	return res, err
}

func (s *sqlStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()
	var (
		rows driver.Rows
		err  error
	)
	if qc, ok := s.parent.(driver.StmtQueryContext); ok {
		rows, err = qc.QueryContext(ctx, args)
	} else {
		var values []driver.Value
		if values, err = positional(args); err == nil {
			rows, err = s.parent.Query(values)
		}
	}
	record(ctx, s.query, args, start, -1, err)

	return rows, err
}

func positional(args []driver.NamedValue) ([]driver.Value, error) {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, errNamedArgs
		}
		values[i] = arg.Value
	}

	return values, nil
}
