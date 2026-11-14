# godebugbar

A web debug toolbar and request profiler for Go. Supports HTTP routers, SQL databases, and custom collectors.

Inspired by the Symfony Web Profiler, `godebugbar` attaches a collapsible debug toolbar to your web pages and API responses during local development. It captures HTTP request details, SQL queries, execution timings, Go runtime metrics, and custom application state per request through `context.Context`.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar@latest
```

## Usage

### Gin and GORM

```go
godebugbar.New(
	ginadapter.Attach(engine),
	gormadapter.Attach(db),
)
```

### Standard `net/http` or `chi`

```go
bar := godebugbar.New()
handler := bar.Middleware(mux)
```

### Custom Collectors

Implement `godebugbar.Collector` and pass it to `WithCollectors` to add a custom badge to the bottom toolbar and a tab to the inspector drawer:

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

Releases follow `CONTRIBUTING.md`. Merging a `release-please` pull request tags `main` (`v0.2.0` for `feat:`, `v0.1.1` for `fix:`) and publishes a GitHub Release.
