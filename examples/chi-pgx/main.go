package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DarkoKlisuric/godebugbar"
	pgxadapter "github.com/DarkoKlisuric/godebugbar/adapter/pgx"
)

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html>
<head><title>{{.Title}}</title><style>body{font-family:system-ui,sans-serif;margin:40px;color:#222}a{margin-right:16px}</style></head>
<body>
<nav><a href="/">N+1 page</a><a href="/joined">Joined page</a><a href="/_debugbar">Profiler</a></nav>
<h1>{{.Title}}</h1>
<ul>{{range .Groups}}<li>{{.Name}}: {{len .Members}} members</li>{{end}}</ul>
</body>
</html>`))

const schema = `
CREATE TABLE IF NOT EXISTS pgx_groups (id serial PRIMARY KEY, name text NOT NULL);
CREATE TABLE IF NOT EXISTS pgx_members (id serial PRIMARY KEY, group_id int NOT NULL REFERENCES pgx_groups(id), name text NOT NULL);
INSERT INTO pgx_groups (name) SELECT unnest(ARRAY['Alpha','Bravo','Charlie','Delta','Echo','Foxtrot']) WHERE NOT EXISTS (SELECT 1 FROM pgx_groups);
INSERT INTO pgx_members (group_id, name) SELECT c.id, c.name || ' ' || n FROM pgx_groups c, generate_series(1, 5) n WHERE NOT EXISTS (SELECT 1 FROM pgx_members);
`

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(envOr("DATABASE_URL", "postgres://godebugbar:godebugbar@localhost:5433/godebugbar?sslmode=disable"))
	if err != nil {
		return err
	}
	config.ConnConfig.Tracer = pgxadapter.NewTracer()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("prepare the database (run docker compose up -d): %w", err)
	}

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		godebugbar.WithHTMLInjection(),
		godebugbar.WithExcludedPaths("/favicon.ico"),
		pgxadapter.Attach(),
	)

	router := chi.NewRouter()
	router.Use(bar.Middleware)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		groups, err := groupsWithMemberQueries(r.Context(), pool)
		render(w, r, "Groups (N+1)", groups, err)
	})
	router.Get("/joined", func(w http.ResponseWriter, r *http.Request) {
		groups, err := groupsJoined(r.Context(), pool)
		render(w, r, "Groups (one join)", groups, err)
	})

	addr := envOr("ADDR", ":8080")
	server := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("open http://localhost" + addr + " and http://localhost" + addr + "/_debugbar")

	return server.ListenAndServe()
}

func groupsWithMemberQueries(ctx context.Context, pool *pgxpool.Pool) ([]group, error) {
	rows, err := pool.Query(ctx, "SELECT id, name FROM pgx_groups ORDER BY name")
	if err != nil {
		return nil, err
	}
	groups, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (group, error) {
		var g group
		return g, row.Scan(&g.ID, &g.Name)
	})
	if err != nil {
		return nil, err
	}
	for i := range groups {
		rows, err := pool.Query(ctx, "SELECT name FROM pgx_members WHERE group_id = $1 ORDER BY id", groups[i].ID)
		if err != nil {
			return nil, err
		}
		if groups[i].Members, err = pgx.CollectRows(rows, pgx.RowTo[string]); err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func groupsJoined(ctx context.Context, pool *pgxpool.Pool) ([]group, error) {
	rows, err := pool.Query(ctx, `SELECT c.id, c.name, array_agg(m.name ORDER BY m.id)
		FROM pgx_groups c JOIN pgx_members m ON m.group_id = c.id
		GROUP BY c.id, c.name ORDER BY c.name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (group, error) {
		var g group
		return g, row.Scan(&g.ID, &g.Name, &g.Members)
	})
}

func render(w http.ResponseWriter, r *http.Request, title string, groups []group, err error) {
	if err != nil {
		godebugbar.FromContext(r.Context()).AddError(err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(w, map[string]any{"Title": title, "Groups": groups}); err != nil {
		godebugbar.FromContext(r.Context()).AddError(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
