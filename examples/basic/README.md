# basic example

A net/http app with no database. It records queries by hand with `godebugbar.FromContext`, logs through `godebugbar.NewSlogHandler` and makes an outgoing call through `godebugbar.NewTransport`.

```bash
go run ./examples/basic
```

- `/` loads members once per group (N+1) and runs one slow query.
- `/fixed` loads all members in one query.
