package gormadapter_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/DarkoKlisuric/godebugbar"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
)

var registerExplainDriver sync.Once

type sampleModel struct {
	ID   uint
	Name string
}

func latestBody(t *testing.T, handler http.Handler) string {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec.Body.String()
}

func TestGormPluginRecordsQueries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		options     []godebugbar.Option
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "request context with values",
			wantPresent: []string{"zebra", "yak", "gorm_plugin_test.go", "N+1 (3×)"},
		},
		{
			name:        "hidden values keep only placeholders",
			options:     []godebugbar.Option{godebugbar.WithHiddenSQLValues()},
			wantPresent: []string{`name = ?`},
			wantAbsent:  []string{"zebra", "yak"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			bar := godebugbar.New(append([]godebugbar.Option{
				godebugbar.WithEnabled(true),
				gormadapter.Attach(db),
			}, tc.options...)...)
			handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tx := db.WithContext(r.Context())
				for range 3 {
					var items []sampleModel
					_ = tx.Where("name = ?", "zebra").Find(&items).Error
				}
				_ = tx.Create(&sampleModel{Name: "yak"}).Error
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			body := latestBody(t, handler)
			var decoded map[string]any
			if err := json.Unmarshal([]byte(body), &decoded); err != nil {
				t.Fatalf("decode: %v", err)
			}
			for _, want := range tc.wantPresent {
				if !strings.Contains(body, want) {
					t.Errorf("expected %q in profile %s", want, body)
				}
			}
			for _, unwanted := range tc.wantAbsent {
				if strings.Contains(body, unwanted) {
					t.Errorf("expected %q not in profile", unwanted)
				}
			}
		})
	}
}

func TestGormPluginStandalone(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Use(gormadapter.Plugin()); err != nil {
		t.Fatalf("Use(Plugin()): %v", err)
	}
	bar := godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithQueryCollector())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var items []sampleModel
		_ = db.WithContext(r.Context()).Where("name = ?", "standalone").Find(&items).Error
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))
	if !strings.Contains(rec.Body.String(), "standalone") {
		t.Fatalf("expected the standalone plugin to record the query, got %s", rec.Body.String())
	}
}

type explainDriver struct{}

func (d *explainDriver) Open(string) (driver.Conn, error) { return &explainConn{}, nil }

type explainConn struct{}

func (c *explainConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }

func (c *explainConn) Close() error { return nil }

func (c *explainConn) Begin() (driver.Tx, error) { return &explainTx{}, nil }

func (c *explainConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &explainTx{}, nil
}

func (c *explainConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "broken") {
		return nil, errors.New("syntax error")
	}
	if strings.HasPrefix(query, "EXPLAIN ") {
		return &planRows{lines: []string{"Seq Scan on sample_models", "Filter: (name = 'zebra')"}}, nil
	}
	return &planRows{}, nil
}

type explainTx struct{}

func (t *explainTx) Commit() error   { return nil }
func (t *explainTx) Rollback() error { return nil }

type planRows struct {
	lines []string
	next  int
}

func (r *planRows) Columns() []string { return []string{"QUERY PLAN"} }
func (r *planRows) Close() error      { return nil }

func (r *planRows) Next(dest []driver.Value) error {
	if r.next >= len(r.lines) {
		return io.EOF
	}
	dest[0] = r.lines[r.next]
	r.next++
	return nil
}

func TestGormPluginExplain(t *testing.T) {
	t.Parallel()

	registerExplainDriver.Do(func() { sql.Register("godebugbar-gorm-explain", &explainDriver{}) })
	sqlDB, err := sql.Open("godebugbar-gorm-explain", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	bar := godebugbar.New(godebugbar.WithEnabled(true), gormadapter.Attach(db))
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var items []sampleModel
		_ = db.WithContext(r.Context()).Where("name = ?", "zebra").Find(&items).Error
		_ = db.WithContext(r.Context()).Raw("SELECT broken").Scan(&items).Error
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/items", nil))
	id := rec.Header().Get("X-Debugbar-ID")

	cases := []struct {
		name       string
		index      string
		wantStatus int
		wantBody   string
	}{
		{name: "select is explained", index: "1", wantStatus: http.StatusOK, wantBody: "Seq Scan on sample_models"},
		{name: "explain error is reported", index: "2", wantStatus: http.StatusUnprocessableEntity, wantBody: "syntax error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			explain := httptest.NewRecorder()
			handler.ServeHTTP(explain, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/"+id+"/queries/"+tc.index+"/explain", nil))
			if explain.Code != tc.wantStatus || !strings.Contains(explain.Body.String(), tc.wantBody) {
				t.Errorf("explain = %d %s", explain.Code, explain.Body.String())
			}
		})
	}
}

func TestGormAttachWithoutDB(t *testing.T) {
	t.Parallel()

	bar := godebugbar.New(godebugbar.WithEnabled(true), gormadapter.Attach(nil))
	if bar == nil {
		t.Fatal("expected a profiler")
	}
}
