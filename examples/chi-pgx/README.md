# chi-pgx example

chi and a pgx v5 pool with `pgxadapter.NewTracer` on PostgreSQL. chi uses the plain `bar.Middleware`.

```bash
cd examples && docker compose up -d
go run ./chi-pgx
```

- `/` loads members once per group (N+1).
- `/joined` loads everything with one join.
