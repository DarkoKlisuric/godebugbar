# gorm adapter

`adapter/gorm` integrates `godebugbar` with `gorm.io/gorm`.

## Usage

Pass `gormadapter.Attach(db)` to `godebugbar.New`:

```go
godebugbar.New(
	ginadapter.Attach(engine),
	gormadapter.Attach(db),
)
```

`Attach` registers a `gorm.Plugin` on `db` that captures every SQL statement (`Query`, `Create`, `Update`, `Delete`, `Row`, `Raw`) executed with a request's `context.Context` (`db.WithContext(ctx)`). Each query's SQL text, bound variables, rows affected, execution duration, caller file and line, and errors are recorded into the request profile.

## Testing

```bash
go test -race ./adapter/gorm/...
go test -race -tags=integration ./adapter/gorm/...
```
