# collector

`internal/collector` implements the built-in default collectors registered by `godebugbar.New`:

- `RequestCollector`: HTTP status, method, route, duration, size, headers, query parameters, and errors.
- `QueryCollector`: SQL statements recorded in `internal/contextstate`, total database duration, and duplicate query detection.
- `RuntimeCollector`: Go heap allocation, system memory, goroutine count, and GC cycles.
