package sqladapter

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/DarkoKlisuric/godebugbar"
)

var errIsolationUnsupported = errors.New("godebugbar: the wrapped driver does not support transaction options")

type sqlConn struct {
	parent driver.Conn
}

func (c *sqlConn) Prepare(query string) (driver.Stmt, error) {
	s, err := c.parent.Prepare(query)
	if err != nil {
		return nil, err
	}

	return &sqlStmt{parent: s, query: query}, nil
}

func (c *sqlConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		s   driver.Stmt
		err error
	)
	if pc, ok := c.parent.(driver.ConnPrepareContext); ok {
		s, err = pc.PrepareContext(ctx, query)
	} else {
		s, err = c.parent.Prepare(query)
	}
	if err != nil {
		return nil, err
	}

	return &sqlStmt{parent: s, query: query}, nil
}

func (c *sqlConn) Close() error {
	return c.parent.Close()
}

func (c *sqlConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *sqlConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if bt, ok := c.parent.(driver.ConnBeginTx); ok {
		return bt.BeginTx(ctx, opts)
	}
	if opts.Isolation != 0 || opts.ReadOnly {
		return nil, errIsolationUnsupported
	}

	return c.parent.Begin()
}

func (c *sqlConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	qc, ok := c.parent.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	start := time.Now()
	rows, err := qc.QueryContext(ctx, query, args)
	record(ctx, query, args, start, -1, err)

	return rows, err
}

func (c *sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ec, ok := c.parent.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	start := time.Now()
	res, err := ec.ExecContext(ctx, query, args)
	record(ctx, query, args, start, rowsAffected(res), err)

	return res, err
}

func (c *sqlConn) Ping(ctx context.Context) error {
	if p, ok := c.parent.(driver.Pinger); ok {
		return p.Ping(ctx)
	}

	return nil
}

func (c *sqlConn) ResetSession(ctx context.Context) error {
	if r, ok := c.parent.(driver.SessionResetter); ok {
		return r.ResetSession(ctx)
	}

	return nil
}

func (c *sqlConn) IsValid() bool {
	if v, ok := c.parent.(driver.Validator); ok {
		return v.IsValid()
	}

	return true
}

func (c *sqlConn) CheckNamedValue(nv *driver.NamedValue) error {
	if checker, ok := c.parent.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nv)
	}

	return driver.ErrSkip
}

func record(ctx context.Context, query string, args []driver.NamedValue, start time.Time, rows int64, err error) {
	request := godebugbar.FromContext(ctx)
	if request == nil || errors.Is(err, driver.ErrSkip) {
		return
	}

	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	request.AddQuery(godebugbar.Query{
		SQL:          query,
		Template:     query,
		Args:         values,
		RowsAffected: max(rows, 0),
		Duration:     time.Since(start),
		StartedAt:    start,
		Err:          err,
	})
}

func rowsAffected(res driver.Result) int64 {
	if res == nil {
		return -1
	}
	n, err := res.RowsAffected()
	if err != nil {
		return -1
	}

	return n
}
