# gorm adapter

`adapter/gorm` integrates `godebugbar` with `gorm.io/gorm`.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar/adapter/gorm@latest
```

## Usage

Pass `gormadapter.Attach(db)` to `godebugbar.New`:

```go
godebugbar.New(
	godebugbar.WithEnabled(true),
	ginadapter.Attach(engine),
	gormadapter.Attach(db),
)
```

`Attach` registers the query collector, installs the plugin on `db` and enables the EXPLAIN button for recorded SELECT statements. When the `*gorm.DB` is built elsewhere (for example in another DI module), register the plugin there and the collector on the profiler:

```go
_ = db.Use(gormadapter.Plugin())
godebugbar.New(godebugbar.WithEnabled(true), godebugbar.WithQueryCollector())
```

Every statement (`Create`, `Query`, `Update`, `Delete`, `Row`, `Raw`) executed with a profiled request's context (`db.WithContext(ctx)`) is recorded: the SQL with values inlined, the placeholder template (used for N+1 detection), bound variables, rows affected, duration, caller `file:line` and error. With `godebugbar.WithHiddenSQLValues()` only the template is kept.

EXPLAIN runs `EXPLAIN <statement>` (never `ANALYZE`) in a read-only transaction, and only for single SELECT or WITH statements.

## Testing

```bash
go test -race ./adapter/gorm/...
go test -race -tags=integration ./adapter/gorm/...
```
