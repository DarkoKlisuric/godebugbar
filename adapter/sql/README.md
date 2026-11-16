# database/sql adapter

`adapter/sql` wraps any `database/sql` driver so that statements run with a profiled request's context are recorded.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar/adapter/sql@latest
```

## Usage

```go
db, err := sqladapter.Open("pgx", dsn)

bar := godebugbar.New(
	godebugbar.WithEnabled(true),
	sqladapter.Attach(),
)
```

`Open` works like `sql.Open` with the named driver wrapped. For a connector use `sql.OpenDB(sqladapter.WrapConnector(connector))`; to register a wrapped driver use `sql.Register("profiled-pgx", sqladapter.Wrap(driver))`. Only context-aware calls (`QueryContext`, `ExecContext`) are recorded, because only they carry the request context. Rows returned by a query are not counted; rows affected by an exec are.

## Testing

```bash
go test -race ./adapter/sql/...
```
