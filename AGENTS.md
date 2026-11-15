This is the code base of `godebugbar`, a web debug toolbar and request profiler for Go applications.

`godebugbar` follows Go idiomatic encapsulation with a minimal public API at the root, internal implementation packages under `internal/`, and loosely coupled adapter packages under `adapter/`:

- **Public API** (`.`) - `Profiler` (`New`, `Middleware`, `Handler`), `Option` functions (`WithEnabled`, `WithAuthorizer`, `WithCORS`, `WithRedacted*`, `WithCollectors`, ...), the per-request `Request` API (`FromContext`, `AddQuery`, `AddError`, `Log`, `StartSpan`), `SlogHandler`, `Transport`, and the `Collector` / `Section` / `Panel` contracts
- **Internal Packages** (`internal/`):
  - `internal/contract/` - Collector, Section, ToolbarItem, ResponseInfo, Panel, Widget and TimelineEvent contracts
  - `internal/contextstate/` - Per-request `Recorder` (queries, errors, logs, spans, outgoing calls, panic, body) stored in `context.Context`
  - `internal/collector/` - Built-in collectors (request, runtime, query, log, HTTP client, timeline)
  - `internal/engine/` - `Engine`, `Session`, ring-buffer `Store`, `Persister`, `ResponseRecorder` and the profiler endpoints
  - `internal/redact/` - Masking of headers, query parameters, URLs and request bodies
  - `internal/ui/` - Embedded renderer, toolbar and dashboard assets
- **Adapters** (`adapter/`): `adapter/gin` (`ginadapter.Attach`), `adapter/gorm` (`gormadapter.Attach`, `gormadapter.Plugin`), `adapter/pgx` (`pgxadapter.NewTracer`, `pgxadapter.Attach`), `adapter/sql` (`sqladapter.Open`, `Wrap`, `WrapConnector`, `Attach`)
- **Codecheck** (`tools/codecheck/`) - AST linter enforcing project code conventions

`Option` is `func(*engine.Engine)`: only packages inside this module can write options, which keeps adapters bound to the engine without a registry.

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
	Name() string
	Collect(ctx context.Context, request *http.Request, response *ResponseInfo) *Section
}
```

When adding a collector or adapter:

1. Built-in collectors belong in `internal/collector/<name>_collector.go` with a matching `_test.go` file.
2. Third-party framework or library adapters belong in `adapter/<name>/` and expose only `Attach(...) godebugbar.Option`, keeping adapter structs unexported.
3. Record per-request events into `context.Context` via `contextstate.RecorderFromContext(ctx)` (or the public `godebugbar.FromContext(ctx)`) during request execution. Never record outside a profiled request.
4. Build the toolbar summary (`ToolbarItem`) and detailed view (`Panel`) in `Collect`.
5. Document every Go package directory in its `README.md`.

### Code Style

`make lint` enforces these rules automatically via `golangci-lint` and `tools/codecheck`:

- **Singular names**: Package, file, and type names use the singular (`adapter`, `query_collector.go`, `QueryCollector`). Do not create catch-all packages such as `util`, `utils`, `common`, `models`, or `types`.
- **One concept per file**: Name every file in `snake_case` after its main type (`Profile` in `profile.go`, `gormPlugin` in `gorm_plugin.go`). Small helper types that belong to that concept live in the same file.
- **Pointer receivers**: Every method uses a pointer receiver (`func (p *Profiler) Middleware(...)`).
- **Comments**: Exported declarations of the public API (the root package and `adapter/*`, outside `internal/`) carry Go doc comments, short and to the point, so pkg.go.dev documents them. Everything else has no comments: internal packages, unexported code, function bodies, tests and tools. `//go:` directives are always allowed. Explain architecture and design in each directory's `README.md`.
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
