# godebugbar

A web debug toolbar and request profiler for Go. Supports HTTP routers, SQL databases, and custom collectors.

Inspired by the Symfony Web Profiler, `godebugbar` serves a standalone profiler UI at `/_debugbar` for JSON APIs and can optionally attach a collapsible bottom debug toolbar to HTML responses. It captures HTTP request details, SQL queries, execution timings, Go runtime metrics, and custom application state per request through `context.Context`.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar@latest
```

## Usage

### Gin and GORM

Adapters automatically register their corresponding collector (`ginadapter.Attach` registers the `Request` collector; `gormadapter.Attach` registers the `Queries` collector):

```go
godebugbar.New(
	ginadapter.Attach(engine),
	gormadapter.Attach(db),
)
```

Open `http://localhost:8080/_debugbar` to inspect all recorded API requests.

### Standard `net/http` or `chi`

```go
bar := godebugbar.New(
	godebugbar.WithRequestCollector(),
	godebugbar.WithRuntimeCollector(),
)
handler := bar.Middleware(mux)
```

### Options

| Option | Description |
| :--- | :--- |
| `WithEnabled(bool)` | Enable or disable profiling (default `true`) |
| `WithRequestCollector()` | Register the built-in HTTP Request collector (auto-registered by `ginadapter.Attach`) |
| `WithQueryCollector()` | Register the built-in SQL Query collector (auto-registered by `gormadapter.Attach`) |
| `WithRuntimeCollector()` | Register the built-in Go Runtime collector |
| `WithCollectors(...Collector)` | Register custom `Collector` implementations |
| `WithDashboard(bool)` | Enable or disable the standalone profiler UI at `GET /_debugbar` (default `true`) |
| `WithHTMLInjection()` | Opt in to auto-injecting the bottom toolbar `<script>` into `text/html` responses |
| `WithCORS(...string)` | Allowed CORS origins for `/_debugbar/*` and `Access-Control-Expose-Headers: X-Debugbar-ID` |
| `WithStorageSize(int)` | Ring-buffer capacity for stored profiles (default `50`) |
| `WithPrefix(string)` | URL prefix for debugbar endpoints and assets (default `"/_debugbar"`) |
| `WithExcludedPrefixes(...string)` | Path prefixes excluded from profiling |
| `WithExcludedPaths(...string)` | Exact paths excluded from profiling |
| `WithExcludedMethods(...string)` | HTTP methods excluded from profiling (default `["OPTIONS"]`) |

### Custom Collectors

Implement `godebugbar.Collector` and pass it to `WithCollectors` to add a custom badge to the toolbar and a tab to the inspector:

```go
type AuthCollector struct{}

func (c *AuthCollector) Name() string {
	return "auth"
}

func (c *AuthCollector) Collect(ctx context.Context, _ *http.Request, _ *godebugbar.ResponseInfo) *godebugbar.Section {
	user := userFromContext(ctx)

	panel := godebugbar.NewPanel("Authentication").
		AddKeyValue("Principal", map[string]string{
			"id":   user.ID,
			"role": user.Role,
		})

	return godebugbar.NewSection(
		godebugbar.ToolbarItem{
			Label: "User",
			Value: user.Email,
			Color: godebugbar.ColorGreen,
		},
		panel,
	)
}
```

Register it alongside your adapters:

```go
godebugbar.New(
	ginadapter.Attach(engine),
	gormadapter.Attach(db),
	godebugbar.WithCollectors(&AuthCollector{}),
)
```

## Development

```bash
make test               # run unit tests with -race
make test-integration   # run PostgreSQL integration tests
make lint               # run golangci-lint and codecheck
make fmt                # format code with gofumpt and goimports
```

## Versions and Releases

Releases follow `CONTRIBUTING.md`. Merging a `release-please` pull request tags `main` (`v0.2.0` for `feat:`, `v0.1.1` for `fix:` / `refactor:`) and publishes a GitHub Release.
