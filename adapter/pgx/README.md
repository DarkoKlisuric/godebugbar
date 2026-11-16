# pgx adapter

`adapter/pgx` records queries run through `github.com/jackc/pgx/v5` with a profiled request's context.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar/adapter/pgx@latest
```

## Usage

```go
config, _ := pgxpool.ParseConfig(dsn)
config.ConnConfig.Tracer = pgxadapter.NewTracer()
pool, _ := pgxpool.NewWithConfig(ctx, config)

bar := godebugbar.New(
	godebugbar.WithEnabled(true),
	pgxadapter.Attach(),
)
```

`Attach` registers the query collector. Queries are recorded only when they run with the request context (`pool.Query(r.Context(), ...)`). To keep an existing tracer such as `tracelog`, combine both with `github.com/jackc/pgx/v5/multitracer`.

## Testing

```bash
go test -race ./adapter/pgx/...
```
