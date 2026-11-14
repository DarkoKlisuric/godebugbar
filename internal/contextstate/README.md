# contextstate

`internal/contextstate` holds the request-scoped `Recorder` stored in `context.Context` while a request is profiled. Adapters, the public `godebugbar.Request` API, `SlogHandler` and `Transport` write into it; collectors read from it.

Each kind of entry (queries, errors, logs, spans, outgoing calls) is capped by `Config.EntryLimit`, and entries over the cap are counted in `Dropped`. `RecordQuery` fills in the template, the caller (`Caller`, skipping godebugbar, ORM and driver frames plus `Config.SourceSkip`) and the start time, and drops values when `Config.HideSQLValues` is set.
