# contextstate

`internal/contextstate` holds the request-scoped `Recorder` stored in `context.Context` during a profiled HTTP request. Internal adapters and built-in collectors use it to record SQL queries, route patterns, and errors without exposing internal recording types on the public `godebugbar` API.
