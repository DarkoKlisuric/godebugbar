# examples

Runnable apps that show godebugbar finding an N+1 query. Each page has a fixed version next to it, so you can compare both in the profiler.

| Example | Stack | Needs |
| :--- | :--- | :--- |
| [`basic`](basic) | net/http, recorded queries, slog, outgoing HTTP call | nothing |
| [`gin-gorm`](gin-gorm) | Gin + GORM | PostgreSQL |
| [`chi-pgx`](chi-pgx) | chi + pgx v5 | PostgreSQL |

```bash
cd examples
docker compose up -d     # PostgreSQL on localhost:5433, only for gin-gorm and chi-pgx
go run ./basic
```

Open `http://localhost:8080`, then the toolbar or `http://localhost:8080/_debugbar`. Set `ADDR` to use another port and `DATABASE_URL` for another database.

This folder is its own Go module, so its dependencies never reach the core module.
