package godebugbar_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/DarkoKlisuric/godebugbar"
)

func ExampleNew() {
	bar := godebugbar.New(godebugbar.WithEnabled(true))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "hello")
	})
	handler := bar.Middleware(mux)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/hello", nil))

	fmt.Println(rec.Code, rec.Body.String())
	fmt.Println("profiled:", rec.Header().Get("X-Debugbar-Id") != "")
	// Output:
	// 200 hello
	// profiled: true
}

func ExampleProfiler_Middleware() {
	bar := godebugbar.New(godebugbar.WithEnabled(true))
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/orders", nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/_debugbar/api/profiles?limit=1", nil))

	var profiles []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Status int    `json:"status"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&profiles)
	fmt.Println(profiles[0].Method, profiles[0].Path, profiles[0].Status)
	// Output: POST /orders 204
}

func ExampleFromContext() {
	bar := godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithQueryCollector())
	handler := bar.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := godebugbar.FromContext(r.Context())
		for id := 1; id <= 3; id++ {
			req.AddQuery(godebugbar.Query{
				SQL:      fmt.Sprintf("SELECT * FROM members WHERE group_id = %d", id),
				Template: "SELECT * FROM members WHERE group_id = $1",
				Duration: time.Millisecond,
			})
		}
		req.Log(slog.LevelWarn, "cache miss")
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/groups", nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/_debugbar/api/profiles?limit=1", nil))

	var profiles []struct {
		Stats struct {
			Queries  int `json:"queries"`
			NPlusOne int `json:"n_plus_one"`
			Logs     int `json:"logs"`
		} `json:"stats"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&profiles)
	stats := profiles[0].Stats
	fmt.Println("queries:", stats.Queries, "n+1:", stats.NPlusOne, "logs:", stats.Logs)
	fmt.Println("outside a request:", godebugbar.FromContext(context.Background()) == nil)
	// Output:
	// queries: 3 n+1: 1 logs: 1
	// outside a request: true
}
