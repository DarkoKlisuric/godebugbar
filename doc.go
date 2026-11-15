// Package godebugbar is a debug toolbar and request profiler for Go web apps.
//
// Wrap a handler with Profiler.Middleware. Each request then gets a profile
// with its status, headers, SQL queries, logs, outgoing calls, timeline and
// runtime metrics, shown at /_debugbar and in an optional page toolbar.
// Profiling is off by default; enable it in development only:
//
//	bar := godebugbar.New(godebugbar.WithEnabled(true))
//	http.ListenAndServe(":8080", bar.Middleware(mux))
package godebugbar
