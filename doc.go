// Package godebugbar is a debug toolbar and request profiler for Go web apps.
// It shows SQL queries with N+1 detection, logs, outgoing HTTP calls and a
// timeline for every request.
//
// Wrap a handler with Profiler.Middleware. Each request then gets a profile,
// shown at /_debugbar and in an optional page toolbar. Adapters for Gin, GORM,
// pgx and database/sql live in their own modules under adapter/.
// Profiling is off by default; enable it in development only:
//
//	bar := godebugbar.New(godebugbar.WithEnabled(true))
//	http.ListenAndServe(":8080", bar.Middleware(mux))
package godebugbar
