# godebugbar

A web debug toolbar and request profiler for Go. Supports HTTP routers, SQL databases, and custom collectors.

Inspired by the Symfony Web Profiler, `godebugbar` attaches a collapsible debug toolbar to your web pages and API responses during local development. It captures HTTP request details, SQL queries (GORM, `database/sql`), Goose migration status, structured logs (`log/slog`), outgoing HTTP calls, execution timelines, and Go runtime metrics per request through `context.Context`.

## Architecture

`godebugbar` is organized into a zero-dependency core package, standard-library collectors, and optional framework/database adapters:

- **Core** (`.`): `Profiler`, `Collector` interface, declarative `Panel` and `Widget` builders, request-scoped `Recorder`, ring-buffer `Store`, and standard `net/http` middleware.
- **Built-in Collectors** (`collector/`): `RequestCollector`, `RuntimeCollector`, `SlogCollector`, `HTTPClientCollector`, `SQLCollector`, and `TimelineCollector`.
- **Adapters** (`adapter/`): Prebuilt integrations for `gin` (`adapter/gin`), `gorm` (`adapter/gorm`), and `goose` (`adapter/goose`).
- **UI** (`ui/`): Embedded Shadow DOM bottom toolbar (`/_debugbar/toolbar.js`) with live `fetch` and `XMLHttpRequest` interception via the `X-Debug-Token` header, plus a standalone profiler view at `/_debugbar/`.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar@latest
```

## Usage

### Gin, GORM, Goose, and `log/slog`

```go
package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
	ginadapter "github.com/DarkoKlisuric/godebugbar/adapter/gin"
	gooseadapter "github.com/DarkoKlisuric/godebugbar/adapter/goose"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
	"github.com/DarkoKlisuric/godebugbar/collector"
)

func setupProfiler(engine *gin.Engine, gormDB *gorm.DB, sqlDB *sql.DB) {
	gormPlugin := gormadapter.NewGormPlugin()
	gooseCollector := gooseadapter.NewGooseCollector(sqlDB, "migrations")
	slogCollector := collector.NewSlogCollector()
	httpClientCollector := collector.NewHTTPClientCollector()

	profiler := godebugbar.NewProfiler(
		godebugbar.WithEnabled(os.Getenv("APP_ENV") == "local"),
		godebugbar.WithCollectors(
			gormPlugin,
			gooseCollector,
			slogCollector,
			httpClientCollector,
		),
	)

	_ = gormDB.Use(gormPlugin)
	http.DefaultTransport = httpClientCollector.WrapRoundTripper(http.DefaultTransport)
	slog.SetDefault(slog.New(slogCollector.WrapHandler(slog.NewTextHandler(os.Stdout, nil))))

	engine.ContextWithFallback = true
	engine.Use(ginadapter.NewMiddleware(profiler))
}
```

### Standard `net/http` or `chi`

```go
profiler := godebugbar.NewProfiler(
	godebugbar.WithEnabled(true),
)

handler := profiler.Middleware(mux)
```

### Custom Collectors

Any type implementing `godebugbar.Collector` adds a new badge to the bottom toolbar and a tab to the inspector drawer:

```go
type AuthCollector struct{}

func (c *AuthCollector) Name() string {
	return "auth"
}

func (c *AuthCollector) Collect(ctx context.Context, request *http.Request, response *godebugbar.ResponseInfo) *godebugbar.Section {
	user := userFromContext(ctx)

	panel := godebugbar.NewPanel("Authentication").
		AddKeyValue("Principal", map[string]any{
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

## Development

```bash
make test               # run unit tests with -race
make test-integration   # run PostgreSQL integration tests
make lint               # run golangci-lint and codecheck
make fmt                # format code with gofumpt and goimports
```

## Versions and Releases

Releases follow `CONTRIBUTING.md`. Merging a `release-please` pull request tags `main` (`v0.2.0` for `feat:`, `v0.1.1` for `fix:`) and publishes a GitHub Release.
