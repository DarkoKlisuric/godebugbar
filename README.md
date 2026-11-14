# godebugbar

A web debug toolbar and request profiler for Go, in the spirit of the Symfony Web Profiler and Laravel Debugbar. Supports `net/http`, Gin, GORM, pgx, any `database/sql` driver, `log/slog`, outgoing HTTP calls and custom collectors.

Every profiled request gets a profile: status, route, headers, body, SQL queries with duplicate, slow and N+1 detection, logs, outgoing HTTP calls, a timeline, panics and Go runtime metrics. Profiles are kept in memory and shown on a full-page profiler at `/_debugbar`, and optionally in a collapsible toolbar at the bottom of every HTML page or SPA.

> **Development only.** Profiles contain request data and SQL values. Profiling is **off by default**; enable it explicitly, and keep it away from production traffic.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar@latest
```

## Quick start (`net/http`, chi, any router)

```go
bar := godebugbar.New(
	godebugbar.WithEnabled(os.Getenv("APP_ENV") == "dev"),
	godebugbar.WithHTMLInjection(), // server-rendered pages get the toolbar
)
http.ListenAndServe(":8080", bar.Middleware(mux))
```

Open `http://localhost:8080/_debugbar`.

## Gin + GORM JSON API with a separate SPA

```go
router := gin.New()
router.Use(gin.Recovery())

godebugbar.New(
	godebugbar.WithEnabled(cfg.Env == "dev"),
	godebugbar.WithCORS("http://localhost:5173"),  // the SPA's origin
	godebugbar.WithPublicURL("http://localhost:8080/_debugbar"),
	godebugbar.WithExcludedPrefixes("/ws", "/events"), // WebSocket and SSE endpoints
	ginadapter.Attach(router), // before any route is registered
	gormadapter.Attach(db),
)

router.GET("/users/:id", showUser) // routes come after Attach
```

`ginadapter.Attach` panics if routes are already registered, because Gin applies middleware only to routes added later. It also sets `router.ContextWithFallback`, so `db.WithContext(c)` with a `*gin.Context` records queries.

In the SPA, load the toolbar in development only, before the app bundle so it can see every `fetch` and `XMLHttpRequest`:

```html
<script src="http://localhost:8080/_debugbar/assets/toolbar.js"
        data-base-url="http://localhost:8080/_debugbar"></script>
```

The toolbar follows every API call that returns an `X-Debugbar-ID` header, keeps a history of the last 50, and links each one to the full profiler.

## Database adapters

| Adapter | Wiring |
| :--- | :--- |
| GORM | `gormadapter.Attach(db)`, or `db.Use(gormadapter.Plugin())` plus `godebugbar.WithQueryCollector()` when the DB is built elsewhere. Adds an EXPLAIN button for SELECT statements. |
| pgx v5 | `config.ConnConfig.Tracer = pgxadapter.NewTracer()` and `pgxadapter.Attach()` |
| `database/sql` | `db, err := sqladapter.Open("pgx", dsn)` and `sqladapter.Attach()` |

Queries are recorded only when they run with the request context (`r.Context()`). Repository wrappers can be skipped when finding the caller with `WithSourceSkip("internal/platform/database")`.

## Recording from your own code

```go
req := godebugbar.FromContext(ctx) // nil-safe: a no-op outside a profiled request
req.AddError(err)
req.Log(slog.LevelWarn, "cache miss", slog.String("key", key))
defer req.StartSpan("render invoice")()
req.AddQuery(godebugbar.Query{SQL: sql, Args: args, Duration: d})
```

`Query` fields: `SQL` (as executed), `Template` (with placeholders; groups repeats for N+1 detection, defaults to `SQL`), `Args`, `RowsAffected`, `Duration`, `StartedAt` (defaults to now minus `Duration`), `Err`, and `Source` (`file:line`, found automatically when empty).

Logs and outgoing calls can be captured automatically:

```go
logger := slog.New(godebugbar.NewSlogHandler(slog.NewJSONHandler(os.Stdout, nil)))
client := &http.Client{Transport: godebugbar.NewTransport(nil)}
```

## Custom collectors

A `Collector` turns a finished request into a toolbar badge and an inspector tab. Panels hold metrics, key/value tables, tables (with per-row colors), code blocks and timelines.

```go
type AuthCollector struct{}

func (c *AuthCollector) Name() string { return "auth" }

func (c *AuthCollector) Collect(ctx context.Context, _ *http.Request, _ *godebugbar.ResponseInfo) *godebugbar.Section {
	user := userFromContext(ctx)
	if user == nil {
		return nil // leave the section out
	}

	panel := godebugbar.NewPanel("Authentication").
		AddKeyValue("Principal", map[string]string{"id": user.ID, "role": user.Role})

	return godebugbar.NewSection(godebugbar.ToolbarItem{Label: "User", Value: user.Role, Color: godebugbar.ColorGreen}, panel)
}

godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithCollectors(&AuthCollector{}))
```

Collector names must be unique; `New` panics when two different collectors share one.

## Options

| Option | Description |
| :--- | :--- |
| `WithEnabled(bool)` | Turn profiling on. **Default `false`.** |
| `WithAuthorizer(func(*http.Request) bool)` | Who may see profiles, the profiler UI and the `X-Debugbar-ID` header |
| `WithCORS(...string)` | Origins allowed to call the profiler API and read `X-Debugbar-ID`; `*` is ignored |
| `WithRedactedHeaders(...string)` | More headers to mask (Authorization, Cookie, Set-Cookie, API keys are masked by default) |
| `WithRedactedQueryParams(...string)` | More query parameters to mask (token, password, api_key… by default) |
| `WithRedactedBodyFields(...string)` | More JSON/form body fields to mask (password, token, secret… by default) |
| `WithHiddenSQLValues()` | Record SQL with placeholders only, never bound values |
| `WithRequestBody(maxBytes)` | Capture request bodies up to `maxBytes`, redacted |
| `WithHTMLInjection()` | Inject the toolbar into `text/html` responses |
| `WithMaxHTMLBuffer(bytes)` | Largest HTML response buffered for injection (default 4 MiB); larger ones stream untouched |
| `WithDashboard(bool)` | Serve the full-page profiler at the prefix (default `true`) |
| `WithServeEndpoints(bool)` | Let the middleware serve the profiler endpoints (default `true`); pair `false` with `Profiler.Handler()` on a private listener |
| `WithPrefix(string)` | URL prefix of the profiler (default `/_debugbar`) |
| `WithPublicURL(string)` | Absolute profiler URL the browser uses, for an SPA or a separate listener |
| `WithStorageSize(int)` | Profiles kept in memory (default `50`) |
| `WithPersistence(path)` | Keep profiles in a JSON file so they survive restarts (hot reload) |
| `WithEntryLimit(int)` | Max queries, logs, spans, errors and calls per request (default `500` each) |
| `WithSlowQueryThreshold(time.Duration)` | Mark queries as slow (default `100ms`) |
| `WithNPlusOneThreshold(int)` | Flag a statement repeated this many times as N+1 (default `3`) |
| `WithSourceSkip(...string)` | Frames to skip when finding a query's caller |
| `WithExcludedPrefixes(...string)` / `WithExcludedPaths(...string)` | Paths not profiled |
| `WithExcludedMethods(...string)` | Methods not profiled, in addition to `OPTIONS` |
| `WithCollectors(...Collector)` | Register custom collectors |
| `WithQueryCollector()` | Register the SQL collector (database adapters do it for you) |
| `WithRequestCollector()` / `WithRuntimeCollector()` | Register built-ins after `WithoutDefaultCollectors()` |
| `WithoutDefaultCollectors()` | Skip the request, runtime, log, HTTP-client and timeline collectors |

## Serving the profiler on a private port

```go
bar := godebugbar.New(
	godebugbar.WithEnabled(true),
	godebugbar.WithServeEndpoints(false),
	godebugbar.WithPublicURL("http://localhost:9999/_debugbar"),
)
go http.ListenAndServe("127.0.0.1:9999", bar.Handler())
http.ListenAndServe(":8080", bar.Middleware(mux))
```

## The UI

- **Toolbar**: one badge per collector, colored by severity; click for a resizable panel. Follows new requests (toggle “Follow”), shows “N new” while a panel is open, keeps a 50-entry history, links to the full profiler, minimizes to a corner button and remembers its state in `localStorage`. Esc closes the panel.
- **Profiler** (`/_debugbar`): request list with filter, live updates that pause on demand and while the tab is hidden, deep links (`/_debugbar?id=<profile id>`), sortable and filterable tables, highlighted slow, duplicate and N+1 queries, EXPLAIN for SELECT statements (GORM), timeline, Copy as cURL and JSON export.

## Development

```bash
make test               # run unit tests with -race
make test-integration   # run PostgreSQL integration tests (DATABASE_URL)
make lint               # run golangci-lint and codecheck
make fmt                # format code with gofumpt and goimports
```

## Versions and Releases

Releases follow `CONTRIBUTING.md`. Merging a `release-please` pull request tags `main` (`v0.2.0` for `feat:`, `v0.1.1` for `fix:` / `refactor:`) and publishes a GitHub Release.
