package godebugbar_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar"
)

func profileJSON(t *testing.T, handler http.Handler, target string) string {
	t.Helper()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_debugbar/api/profiles/latest", nil))

	return rec.Body.String()
}

func TestFromContextIsNilSafe(t *testing.T) {
	t.Parallel()

	r := godebugbar.FromContext(context.Background())
	if r != nil {
		t.Fatalf("expected nil outside a profiled request")
	}
	r.AddError(errors.New("ignored"))
	r.AddQuery(godebugbar.Query{SQL: "SELECT 1"})
	r.Log(slog.LevelInfo, "ignored")
	r.SetRoute("/ignored")
	r.StartSpan("ignored")()
}

func TestRequestRecording(t *testing.T) {
	t.Parallel()

	bar := godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithQueryCollector())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := godebugbar.FromContext(r.Context())
		req.SetRoute("/reports/{id}")
		req.AddError(errors.New("cache miss"))
		req.Log(slog.LevelWarn, "slow upstream", slog.Int("attempt", 2))
		end := req.StartSpan("render report")
		req.AddQuery(godebugbar.Query{
			SQL:      "SELECT * FROM reports WHERE id = 9",
			Template: "SELECT * FROM reports WHERE id = $1",
			Args:     []any{9},
			Duration: 150 * time.Millisecond,
		})
		end()
		w.WriteHeader(http.StatusOK)
	}))

	body := profileJSON(t, handler, "/reports/9")
	for _, want := range []string{`"route":"/reports/{id}"`, "cache miss", "slow upstream", "attempt=2", "render report", "SLOW", "request_test.go"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in %s", want, body)
		}
	}
}

func TestSlogHandler(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	logger := slog.New(godebugbar.NewSlogHandler(slog.NewTextHandler(&out, &slog.HandlerOptions{Level: slog.LevelWarn})))
	bar := godebugbar.New(godebugbar.WithEnabled(true))
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scoped := logger.With("group", "alpha").WithGroup("job")
		scoped.InfoContext(r.Context(), "queued", "id", 7)
		scoped.ErrorContext(r.Context(), "failed", "id", 8)
		w.WriteHeader(http.StatusOK)
	}))

	body := profileJSON(t, handler, "/jobs")
	for _, want := range []string{"queued", "failed", "group=alpha", "job.id=7", `"color":"red"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in profile %s", want, body)
		}
	}
	if strings.Contains(out.String(), "queued") || !strings.Contains(out.String(), "failed") {
		t.Fatalf("expected the wrapped handler to keep its own level, got %q", out.String())
	}

	nilNext := godebugbar.NewSlogHandler(nil)
	if nilNext.Enabled(context.Background(), slog.LevelError) {
		t.Fatalf("expected a nil-wrapped handler to be disabled outside a request")
	}
}

func TestTransport(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(upstream.Close)

	client := &http.Client{Transport: godebugbar.NewTransport(nil)}
	bar := godebugbar.New(godebugbar.WithEnabled(true))
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream.URL+"/brew?api_key=SECRET-KEY", nil)
		if res, err := client.Do(req); err == nil {
			_ = res.Body.Close()
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := profileJSON(t, handler, "/coffee")
	if !strings.Contains(body, "/brew") || !strings.Contains(body, `"418"`) || strings.Contains(body, "SECRET-KEY") {
		t.Fatalf("expected the outgoing call with a redacted key, got %s", body)
	}
}
