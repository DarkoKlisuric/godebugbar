This is the code base of `godebugbar`, a web debug toolbar and request profiler for Go applications.

`godebugbar` is structured as a **zero-dependency core package with loosely coupled adapter packages**:

- **Core** (`.`) - `Profiler`, `Collector` interface, built-in collectors (`RequestCollector`, `QueryCollector`, `RuntimeCollector`), `Panel` and `Widget` builders, request `Recorder`, ring-buffer `Store`, `net/http` middleware, and embedded Shadow DOM toolbar (`assets/toolbar.js`)
- **Gin Adapter** (`adapter/gin/`) - `ginadapter.Attach(engine)` wiring Gin middleware, route patterns, status codes, and errors
- **GORM Adapter** (`adapter/gorm/`) - `gormadapter.Attach(db)` registering a `gorm.Plugin` that captures SQL queries, variables, rows affected, duration, caller location, and errors
- **Codecheck** (`tools/codecheck/`) - AST linter enforcing project code conventions

**Key architectural principle**: The root package depends only on the Go standard library. Third-party integrations live in `adapter/<name>/` and return `godebugbar.Option` functions (`Attach`), keeping bootstrapping to a single `godebugbar.New(...)` call.

### Running Tests

Run all unit tests across the repository:

```bash
make test
go test -race ./...
```

Run tests in a specific package or filter by test name:

```bash
go test -race ./adapter/gin
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

1. Create `<name>_collector.go` (or `<name>_plugin.go` / `<name>_adapter.go`) and its matching `_test.go` file.
2. Expose an `Attach(...) godebugbar.Option` function for framework or library adapters so callers wire it inside `godebugbar.New(...)`.
3. Record per-request events into `context.Context` via `RecorderFromContext(ctx)` during request execution.
4. Build the toolbar summary (`ToolbarItem`) and detailed view (`Panel`) in `Collect`.
5. Document the package in its `README.md`.

### Code Style

`make lint` enforces these rules automatically via `golangci-lint` and `tools/codecheck`:

- **Singular names**: Package, file, and type names use the singular (`adapter`, `query_collector.go`, `QueryCollector`). Do not create catch-all packages such as `util`, `utils`, `common`, `models`, or `types`.
- **One concept per file**: Name every file in `snake_case` after its main type (`Profile` in `profile.go`, `GormPlugin` in `gorm_plugin.go`). Small helper types that belong to that concept live in the same file.
- **Pointer receivers**: Every method uses a pointer receiver (`func (p *Profiler) Middleware(...)`).
- **No comments in Go code**: Do not write comments in `.go` files (including `_test.go` files). Only `//go:` compiler directives (`//go:embed`, `//go:generate`, `//go:build`) are permitted. Explain architecture, usage, and design in each directory's `README.md`.
- **Context propagation**: Functions performing I/O or recording request state take `ctx context.Context` as their first parameter. Never store a `context.Context` inside a struct.

### Testing Conventions

- Place each file's tests in `<file>_test.go` in the same directory.
- Use table-driven tests with `t.Parallel()` whenever testing multiple cases:

```go
func TestStatusColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   BadgeColor
	}{
		{name: "200 ok is green", status: 200, want: ColorGreen},
		{name: "404 not found is yellow", status: 404, want: ColorYellow},
		{name: "500 internal error is red", status: 500, want: ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := StatusColor(tt.status); got != tt.want {
				t.Errorf("StatusColor(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
```

### Git and Pull Requests

- Follow `CONTRIBUTING.md`: branch off `main` (`feat/<short-name>`, `fix/<short-name>`, `chore/<short-name>`), open a pull request into `main`, and merge with **Create a merge commit** once CI is green.
- Write Conventional Commit messages without `Co-Authored-By` or AI attribution lines.
