# engine

Internal profiling engine.

- `Engine`: configuration, the `net/http` middleware, collector runs and the profiler endpoints (`Handler`), guarded by the authorizer and CORS rules.
- `Session`: one profiled request, shared by the middleware and the Gin adapter (`Begin`, `PrepareHeaders`, `RecordPanic`, `End`, `Snippet`).
- `ResponseRecorder`: wraps the response writer, supports `Flush`, `Hijack` and `Unwrap`, and buffers only injectable HTML up to a size cap. `InjectSnippet` places the toolbar before the last `</body>`.
- `ProfileStats`: per-request numbers (queries, DB time, slow, duplicate, N+1, failed, errors, logs, calls, panic) stored on every profile and summary.
- `ProfileFilter`: the list endpoint's search (`q`, `method`, `status`, `min_ms`, `has`, `since`, `limit`).
- `Store`: fixed-size ring of profiles with summaries for the list endpoint; `Persister` optionally mirrors it to a JSON file.
- `BodyCapture`: copies up to N bytes of the request body as the handler reads it.
