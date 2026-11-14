# collector

`internal/collector` implements the built-in collectors:

- `RequestCollector`: status, method, route, duration, size, redacted headers and query parameters, captured request body, errors and panics. Default.
- `RuntimeCollector`: heap, total memory, goroutines and GC cycles, read with `runtime/metrics` (no stop-the-world). Default.
- `LogCollector`, `HTTPCollector`, `TimelineCollector`: logs, outgoing HTTP calls and a waterfall of spans, queries and calls. Default; each returns no section when the request has nothing to show.
- `QueryCollector`: SQL statements, total time, slow queries, exact duplicates and N+1 suspects (statements sharing a template), with per-row colors. Registered by the database adapters or `WithQueryCollector`.
