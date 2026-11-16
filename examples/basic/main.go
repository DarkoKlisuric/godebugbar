package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/DarkoKlisuric/godebugbar"
)

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html>
<head><title>{{.Title}}</title><style>body{font-family:system-ui,sans-serif;margin:40px;color:#222}a{margin-right:16px}</style></head>
<body>
<nav><a href="/">N+1 page</a><a href="/fixed">Fixed page</a><a href="/_debugbar">Profiler</a></nav>
<h1>{{.Title}}</h1>
<ul>{{range .Groups}}<li>{{.}}</li>{{end}}</ul>
</body>
</html>`))

var groups = []string{"Alpha", "Bravo", "Charlie", "Delta"}

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr := envOr("ADDR", ":8080")
	logger := slog.New(godebugbar.NewSlogHandler(slog.NewTextHandler(os.Stderr, nil)))
	client := &http.Client{Transport: godebugbar.NewTransport(nil), Timeout: 2 * time.Second}

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		godebugbar.WithHTMLInjection(),
		godebugbar.WithQueryCollector(),
		godebugbar.WithExcludedPaths("/health", "/favicon.ico"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := godebugbar.FromContext(ctx)
		req.AddQuery(godebugbar.Query{SQL: "SELECT id, name FROM groups ORDER BY name", RowsAffected: int64(len(groups)), Duration: 3 * time.Millisecond})
		for id := range groups {
			req.AddQuery(godebugbar.Query{
				SQL:          fmt.Sprintf("SELECT * FROM members WHERE group_id = %d", id+1),
				Template:     "SELECT * FROM members WHERE group_id = $1",
				Args:         []any{id + 1},
				RowsAffected: 5,
				Duration:     2 * time.Millisecond,
			})
		}
		req.AddQuery(godebugbar.Query{SQL: "SELECT count(*) FROM missions WHERE finished = false", RowsAffected: 1, Duration: 140 * time.Millisecond})
		logger.InfoContext(ctx, "groups loaded", "count", len(groups))
		logger.WarnContext(ctx, "cache miss", "key", "groups:active")
		checkHealth(ctx, client, addr)
		render(w, req, "Groups (N+1)")
	})
	mux.HandleFunc("GET /fixed", func(w http.ResponseWriter, r *http.Request) {
		req := godebugbar.FromContext(r.Context())
		req.AddQuery(godebugbar.Query{SQL: "SELECT id, name FROM groups ORDER BY name", RowsAffected: int64(len(groups)), Duration: 3 * time.Millisecond})
		req.AddQuery(godebugbar.Query{SQL: "SELECT * FROM members WHERE group_id = ANY($1)", Args: []any{"{1,2,3,4}"}, RowsAffected: 20, Duration: 3 * time.Millisecond})
		render(w, req, "Groups (one query for members)")
	})

	server := &http.Server{Addr: addr, Handler: bar.Middleware(mux), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("open http://localhost" + addr + " and http://localhost" + addr + "/_debugbar")

	return server.ListenAndServe()
}

func render(w http.ResponseWriter, req *godebugbar.Request, title string) {
	defer req.StartSpan("render")()
	time.Sleep(10 * time.Millisecond)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(w, map[string]any{"Title": title, "Groups": groups}); err != nil {
		req.AddError(err)
	}
}

func checkHealth(ctx context.Context, client *http.Client, addr string) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost"+addr+"/health", nil)
	if err != nil {
		return
	}
	response, err := client.Do(request)
	if err != nil {
		godebugbar.FromContext(ctx).AddError(err)
		return
	}
	_ = response.Body.Close()
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
