# redact

`internal/redact` masks secrets before they are stored in a profile: request and response headers, query parameters, outgoing request URLs, and JSON or form request bodies. `New` starts with the defaults in `DefaultHeaders`, `DefaultQueryParams` and `DefaultBodyFields`; the `godebugbar.WithRedacted*` options add more.
