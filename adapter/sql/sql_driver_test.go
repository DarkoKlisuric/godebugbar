package sqladapter_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/DarkoKlisuric/godebugbar"
	sqladapter "github.com/DarkoKlisuric/godebugbar/adapter/sql"
)

var registerOnce sync.Once

type fakeDriver struct{ contextAware bool }

func (d *fakeDriver) Open(string) (driver.Conn, error) {
	if d.contextAware {
		return &contextConn{}, nil
	}
	return &plainConn{}, nil
}

type plainConn struct{}

func (c *plainConn) Prepare(query string) (driver.Stmt, error) {
	if strings.Contains(query, "fail") {
		return nil, errors.New("prepare failed")
	}
	return &plainStmt{}, nil
}

func (c *plainConn) Close() error { return nil }

func (c *plainConn) Begin() (driver.Tx, error) { return &plainTx{}, nil }

type plainTx struct{}

func (*plainTx) Commit() error   { return nil }
func (*plainTx) Rollback() error { return nil }

type plainStmt struct{}

func (s *plainStmt) Close() error  { return nil }
func (s *plainStmt) NumInput() int { return -1 }

func (s *plainStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(2), nil }

func (s *plainStmt) Query([]driver.Value) (driver.Rows, error) { return &emptyRows{}, nil }

type contextConn struct{ plainConn }

func (c *contextConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &emptyRows{}, nil
}

func (c *contextConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "fail") {
		return nil, errors.New("exec failed")
	}
	return driver.RowsAffected(5), nil
}

type emptyRows struct{}

func (r *emptyRows) Columns() []string         { return []string{"n"} }
func (r *emptyRows) Close() error              { return nil }
func (r *emptyRows) Next([]driver.Value) error { return io.EOF }

func TestSQLAdapterRecordsQueries(t *testing.T) {
	t.Parallel()

	registerOnce.Do(func() {
		sql.Register("godebugbar-fake-plain", &fakeDriver{})
		sql.Register("godebugbar-fake-context", &fakeDriver{contextAware: true})
	})

	cases := []struct {
		name   string
		driver string
		want   []string
	}{
		{name: "driver without context methods", driver: "godebugbar-fake-plain", want: []string{"SELECT plain_marker", "UPDATE plain_marker", `"2"`}},
		{name: "driver with context methods", driver: "godebugbar-fake-context", want: []string{"SELECT plain_marker", "UPDATE plain_marker", "exec failed"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, err := sqladapter.Open(tc.driver, "dsn")
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			t.Cleanup(func() { _ = db.Close() })

			bar := godebugbar.New(godebugbar.WithEnabled(true), sqladapter.Attach())
			handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rows, err := db.QueryContext(r.Context(), "SELECT plain_marker WHERE id = ?", 7)
				if err == nil {
					_ = rows.Close()
				}
				_, _ = db.ExecContext(r.Context(), "UPDATE plain_marker SET n = ?", 1)
				_, _ = db.ExecContext(r.Context(), "UPDATE fail")
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))

			for _, want := range tc.want {
				if !strings.Contains(rec.Body.String(), want) {
					t.Errorf("expected %q in %s", want, rec.Body.String())
				}
			}
		})
	}
}

func TestSQLAdapterSkipsUnprofiledContext(t *testing.T) {
	t.Parallel()

	connector, err := sqladapter.Wrap(&fakeDriver{contextAware: true}).OpenConnector("dsn")
	if err != nil {
		t.Fatalf("OpenConnector: %v", err)
	}
	db := sql.OpenDB(connector)
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(t.Context(), "UPDATE x"); err != nil {
		t.Fatalf("exec outside a profiled request: %v", err)
	}
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

type fakeConnector struct{ fail bool }

func (c *fakeConnector) Connect(context.Context) (driver.Conn, error) {
	if c.fail {
		return nil, errors.New("connect failed")
	}
	return &contextConn{}, nil
}

func (c *fakeConnector) Driver() driver.Driver { return &fakeDriver{contextAware: true} }

type failingDriver struct{}

func (d *failingDriver) Open(string) (driver.Conn, error) { return nil, errors.New("open failed") }

func TestWrapConnectorRecordsQueries(t *testing.T) {
	t.Parallel()

	db := sql.OpenDB(sqladapter.WrapConnector(&fakeConnector{}))
	t.Cleanup(func() { _ = db.Close() })
	if _, ok := db.Driver().(*sqladapter.SQLDriver); !ok {
		t.Fatalf("Driver() = %T, want *sqladapter.SQLDriver", db.Driver())
	}

	bar := godebugbar.New(godebugbar.WithEnabled(true), sqladapter.Attach())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = db.ExecContext(r.Context(), "DELETE FROM connector_marker")
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))

	if !strings.Contains(rec.Body.String(), "connector_marker") {
		t.Errorf("expected the query in %s", rec.Body.String())
	}
}

func TestSQLDriverErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "open with an unknown driver", run: func() error {
			_, err := sqladapter.Open("godebugbar-unknown", "dsn")
			return err
		}},
		{name: "wrapped driver fails to open", run: func() error {
			_, err := sqladapter.Wrap(&failingDriver{}).Open("dsn")
			return err
		}},
		{name: "wrapped connector fails to connect", run: func() error {
			_, err := sqladapter.WrapConnector(&fakeConnector{fail: true}).Connect(t.Context())
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.run(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
