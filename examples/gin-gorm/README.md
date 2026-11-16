# gin-gorm example

Gin and GORM on PostgreSQL with `ginadapter.Attach` and `gormadapter.Attach`.

```bash
cd examples && docker compose up -d
go run ./gin-gorm
```

- `/` loads members once per group (N+1).
- `/preload` uses `Preload("Members")`: two queries in total.

Open a query in the profiler's Database tab and press EXPLAIN to see the PostgreSQL plan.
