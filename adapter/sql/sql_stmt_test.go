package sqladapter_test

import (
	"context"
	"database/sql/driver"
	"testing"
)

func TestSQLStmtDelegates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		run     func(ctx context.Context, s driver.Stmt) error
		wantErr bool
	}{
		{name: "exec", run: func(_ context.Context, s driver.Stmt) error {
			_, err := s.Exec(nil)
			return err
		}},
		{name: "query", run: func(_ context.Context, s driver.Stmt) error {
			_, err := s.Query(nil)
			return err
		}},
		{name: "exec with positional args", run: func(ctx context.Context, s driver.Stmt) error {
			_, err := s.(driver.StmtExecContext).ExecContext(ctx, []driver.NamedValue{{Ordinal: 1, Value: 1}})
			return err
		}},
		{name: "exec with named args", wantErr: true, run: func(ctx context.Context, s driver.Stmt) error {
			_, err := s.(driver.StmtExecContext).ExecContext(ctx, []driver.NamedValue{{Name: "id", Ordinal: 1, Value: 1}})
			return err
		}},
		{name: "query with named args", wantErr: true, run: func(ctx context.Context, s driver.Stmt) error {
			_, err := s.(driver.StmtQueryContext).QueryContext(ctx, []driver.NamedValue{{Name: "id", Ordinal: 1, Value: 1}})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := openConn(t, &fakeDriver{})
			s, err := c.Prepare("SELECT 1")
			if err != nil {
				t.Fatalf("Prepare: %v", err)
			}
			t.Cleanup(func() { _ = s.Close() })
			if s.NumInput() != -1 {
				t.Errorf("NumInput = %d, want -1", s.NumInput())
			}
			if err := tt.run(t.Context(), s); (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
