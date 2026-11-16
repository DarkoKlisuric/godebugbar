package sqladapter_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	sqladapter "github.com/DarkoKlisuric/godebugbar/adapter/sql"
)

type fullConn struct{ contextConn }

func (c *fullConn) PrepareContext(_ context.Context, query string) (driver.Stmt, error) {
	return c.Prepare(query)
}

func (c *fullConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &plainTx{}, nil
}

func (c *fullConn) Ping(context.Context) error { return errors.New("ping failed") }

func (c *fullConn) ResetSession(context.Context) error { return errors.New("reset failed") }

func (c *fullConn) IsValid() bool { return false }

func (c *fullConn) CheckNamedValue(*driver.NamedValue) error { return errors.New("bad value") }

type fullDriver struct{}

func (d *fullDriver) Open(string) (driver.Conn, error) { return &fullConn{}, nil }

type connMethods interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
	driver.Pinger
	driver.SessionResetter
	driver.Validator
	driver.NamedValueChecker
}

func openConn(t *testing.T, d driver.Driver) connMethods {
	t.Helper()

	conn, err := sqladapter.Wrap(d).Open("dsn")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	c, ok := conn.(connMethods)
	if !ok {
		t.Fatalf("conn %T lacks driver interfaces", conn)
	}

	return c
}

func TestSQLConnDelegates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		driver  driver.Driver
		check   func(ctx context.Context, c connMethods) error
		wantErr bool
	}{
		{name: "prepare", driver: &fakeDriver{}, check: func(_ context.Context, c connMethods) error {
			_, err := c.Prepare("SELECT 1")
			return err
		}},
		{name: "prepare error", driver: &fakeDriver{}, wantErr: true, check: func(_ context.Context, c connMethods) error {
			_, err := c.Prepare("SELECT fail")
			return err
		}},
		{name: "prepare with context", driver: &fullDriver{}, check: func(ctx context.Context, c connMethods) error {
			_, err := c.PrepareContext(ctx, "SELECT 1")
			return err
		}},
		{name: "prepare with context error", driver: &fullDriver{}, wantErr: true, check: func(ctx context.Context, c connMethods) error {
			_, err := c.PrepareContext(ctx, "SELECT fail")
			return err
		}},
		{name: "begin", driver: &fakeDriver{}, check: func(_ context.Context, c connMethods) error {
			_, err := c.Begin()
			return err
		}},
		{name: "read-only transaction without driver support", driver: &fakeDriver{}, wantErr: true, check: func(ctx context.Context, c connMethods) error {
			_, err := c.BeginTx(ctx, driver.TxOptions{ReadOnly: true})
			return err
		}},
		{name: "read-only transaction with driver support", driver: &fullDriver{}, check: func(ctx context.Context, c connMethods) error {
			_, err := c.BeginTx(ctx, driver.TxOptions{ReadOnly: true})
			return err
		}},
		{name: "ping without driver support", driver: &fakeDriver{}, check: func(ctx context.Context, c connMethods) error {
			return c.Ping(ctx)
		}},
		{name: "ping is passed through", driver: &fullDriver{}, wantErr: true, check: func(ctx context.Context, c connMethods) error {
			return c.Ping(ctx)
		}},
		{name: "reset without driver support", driver: &fakeDriver{}, check: func(ctx context.Context, c connMethods) error {
			return c.ResetSession(ctx)
		}},
		{name: "reset is passed through", driver: &fullDriver{}, wantErr: true, check: func(ctx context.Context, c connMethods) error {
			return c.ResetSession(ctx)
		}},
		{name: "valid without driver support", driver: &fakeDriver{}, check: func(_ context.Context, c connMethods) error {
			if !c.IsValid() {
				return errors.New("want valid")
			}
			return nil
		}},
		{name: "validity is passed through", driver: &fullDriver{}, check: func(_ context.Context, c connMethods) error {
			if c.IsValid() {
				return errors.New("want invalid")
			}
			return nil
		}},
		{name: "value check without driver support", driver: &fakeDriver{}, check: func(_ context.Context, c connMethods) error {
			if err := c.CheckNamedValue(&driver.NamedValue{}); !errors.Is(err, driver.ErrSkip) {
				return errors.New("want driver.ErrSkip")
			}
			return nil
		}},
		{name: "value check is passed through", driver: &fullDriver{}, wantErr: true, check: func(_ context.Context, c connMethods) error {
			return c.CheckNamedValue(&driver.NamedValue{})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := openConn(t, tt.driver)
			t.Cleanup(func() { _ = c.Close() })
			if err := tt.check(t.Context(), c); (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
