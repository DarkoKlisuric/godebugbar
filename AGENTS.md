This is the code base of `godebugbar`, a web debug toolbar and request profiler for Go applications.

`godebugbar` follows Go idiomatic encapsulation with a minimal public API at the root, internal implementation packages under `internal/`, and loosely coupled adapter packages under `adapter/`:

- **Public API** (`.`) - `Profiler` (`New`, `Middleware`), `Option` functions (`WithEnabled`, `WithDefaultCollectors`, `WithStorageSize`, `WithPrefix`, `WithCollectors`, `WithExcludedPrefixes`), and custom `Collector` / `Section` / `Panel` contracts
- **Internal Packages** (`internal/`):
  - `internal/contract/` - Collector, Section, ToolbarItem, ResponseInfo, Panel, and Widget contracts
  - `internal/contextstate/` - Per-request `Recorder` and `QueryEntry` stored in `context.Context`
  - `internal/collector/` - Built-in collectors (`RequestCollector`, `QueryCollector`, `RuntimeCollector`)
  - `internal/engine/` - Profiling `Engine`, ring-buffer `Store`, `Profile`, `ResponseRecorder`, and debugbar HTTP endpoints
  - `internal/ui/` - Embedded Shadow DOM toolbar script (`assets/toolbar.js`)
- **Gin Adapter** (`adapter/gin/`) - `ginadapter.Attach(engine)` wiring Gin middleware, route patterns, status codes, and errors
- **GORM Adapter** (`adapter/gorm/`) - `gormadapter.Attach(db)` registering a `gorm.Plugin` that captures SQL queries, variables, rows affected, duration, and errors
- **Codecheck** (`tools/codecheck/`) - AST linter enforcing project code conventions

**Key architectural principle**: Only types and functions meant to be called by external users are exported in the root package and `adapter/*` packages. All internal mechanics live in `internal/*` so external callers cannot depend on or invoke internal methods.

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
	Collect(r *http.Request, res ResponseInfo) Section
}
```

When adding a collector or adapter:

1. Built-in collectors belong in `internal/collector/<name>_collector.go` with a matching `_test.go` file.
2. Third-party framework or library adapters belong in `adapter/<name>/` and expose only `Attach(...) godebugbar.Option`, keeping adapter structs unexported.
3. Record per-request events into `context.Context` via `contextstate.RecorderFromContext(ctx)` during request execution.
4. Build the toolbar summary (`ToolbarItem`) and detailed view (`Panel`) in `Collect`.
5. Document every Go package directory in its `README.md`.

### Code Style

`make lint` enforces these rules automatically via `golangci-lint` and `tools/codecheck`:

- **Singular names**: Package, file, and type names use the singular (`adapter`, `query_collector.go`, `QueryCollector`). Do not create catch-all packages such as `util`, `utils`, `common`, `models`, or `types`.
- **One concept per file**: Name every file in `snake_case` after its main type (`Profile` in `profile.go`, `gormPlugin` in `gorm_plugin.go`). Small helper types that belong to that concept live in the same file.
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

- Follow `CONTRIBUTING.md`: branch off `main` (`feat/<short-name>`, `fix/<short-name>`, `refactor/<short-name>`, `chore/<short-name>`), open a pull request into `main`, and merge with **Create a merge commit** once CI is green.
- Write Conventional Commit messages without `Co-Authored-By` or AI attribution lines.
