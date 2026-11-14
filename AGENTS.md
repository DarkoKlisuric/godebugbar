This is the code base of `godebugbar`, a web debug toolbar and request profiler for Go applications.

`godebugbar` is structured as a **zero-dependency core package with loosely coupled collector and adapter packages**:

- **Core** (`.`) - `Profiler`, `Collector` interface, `Panel` and `Widget` builders, request `Recorder`, ring-buffer `Store`, and `net/http` middleware
- **Collectors** (`collector/`) - Standard-library data collectors (`RequestCollector`, `RuntimeCollector`, `SlogCollector`, `HTTPClientCollector`, `SQLCollector`, `TimelineCollector`)
- **Gin Adapter** (`adapter/gin/`) - Gin middleware capturing matched route patterns, parameters, status codes, and errors
- **GORM Adapter** (`adapter/gorm/`) - `gorm.Plugin` capturing SQL queries, bindings, rows affected, duration, caller location, and duplicate query warnings
- **Goose Adapter** (`adapter/goose/`) - Migration collector inspecting `goose_db_version` and migration files
- **UI** (`ui/`) - Embedded Shadow DOM bottom debug bar and `/_debugbar` dashboard (`//go:embed`)
- **Codecheck** (`tools/codecheck/`) - AST linter enforcing project code conventions

**Key architectural principle**: The root package, `collector/`, and `ui/` depend only on the Go standard library. Third-party integrations live in `adapter/<name>/` and depend on the core interfaces, never the reverse.

### Running Tests

Run all unit tests across the repository:

```bash
make test
go test -race ./...
```

Run tests in a specific package or filter by test name:

```bash
go test -race ./collector
go test -race ./adapter/gorm -run TestGormPlugin
```

Run PostgreSQL integration tests (requires a running PostgreSQL instance on `DATABASE_URL`):

```bash
make test-integration
go test -race -tags=integration ./...
```

### Linting and Formatting

Run both `golangci-lint` and `tools/codecheck`:

```bash
make lint
go run ./tools/codecheck .
```

Format code with `gofumpt` and `goimports`:

```bash
make fmt
```

### Adding a New Collector or Adapter

Every data source implements the `Collector` interface:

```go
type Collector interface {
	Name() string
	Collect(ctx context.Context, request *http.Request, response *ResponseInfo) *Section
}
```

When adding a collector or adapter:

1. Create `<name>_collector.go` (or `<name>_plugin.go` / `<name>_middleware.go`) and its matching `<name>_collector_test.go`.
2. Record per-request events into `context.Context` via `RecorderFromContext(ctx)` during request execution.
3. Build the toolbar summary (`ToolbarItem`) and detailed view (`Panel`) in `Collect`.
4. Document the package in its `README.md`.

### Code Style

`make lint` enforces these rules automatically via `golangci-lint` and `tools/codecheck`:

- **Singular names**: Package, file, and type names use the singular (`collector`, `adapter`, `query_collector.go`, `QueryCollector`). Do not create catch-all packages such as `util`, `utils`, `common`, `models`, or `types`.
- **One concept per file**: Name every file in `snake_case` after its main type (`Profile` in `profile.go`, `GormPlugin` in `gorm_plugin.go`). Small helper types that belong to that concept live in the same file.
- **Pointer receivers**: Every method uses a pointer receiver (`func (p *Profiler) Middleware(...)`).
- **No comments in Go code**: Do not write comments in `.go` files (including `_test.go` files). Only `//go:` compiler directives (`//go:embed`, `//go:generate`, `//go:build`) are permitted. Explain architecture, usage, and design in each directory's `README.md`.
- **Context propagation**: Functions performing I/O or recording request state take `ctx context.Context` as their first parameter. Never store a `context.Context` inside a struct.

### Testing Conventions

- Place each file's tests in `<file>_test.go` in the same directory.
- Use table-driven tests with `t.Parallel()` whenever testing multiple cases:

```go
func TestBadgeColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   BadgeColor
	}{
		{name: "200 is green", status: 200, want: ColorGreen},
		{name: "404 is yellow", status: 404, want: ColorYellow},
		{name: "500 is red", status: 500, want: ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := statusColor(tt.status); got != tt.want {
				t.Errorf("statusColor(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
```

### Git and Pull Requests

- Follow `CONTRIBUTING.md`: branch off `main` (`feat/<short-name>`, `fix/<short-name>`, `chore/<short-name>`), open a pull request into `main`, and merge with **Create a merge commit** once CI is green.
- Write Conventional Commit messages without `Co-Authored-By` or AI attribution lines.
