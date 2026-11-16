# gin adapter

`adapter/gin` integrates `godebugbar` with `github.com/gin-gonic/gin`.

## Install

```bash
go get github.com/DarkoKlisuric/godebugbar/adapter/gin@latest
```

## Usage

Pass `ginadapter.Attach(engine)` to `godebugbar.New` **before registering routes**:

```go
engine := gin.New()
engine.Use(gin.Recovery())

godebugbar.New(
	godebugbar.WithEnabled(true),
	ginadapter.Attach(engine),
)

engine.GET("/users/:id", showUser)
```

Gin applies middleware only to routes added after `Use`, so `Attach` panics when the engine already has routes. When the profiler is enabled, `Attach`:

- registers the profiler middleware and serves `/_debugbar` and its API on the engine,
- sets `engine.ContextWithFallback = true`, so a `*gin.Context` passed as `context.Context` (for example `db.WithContext(c)`) reaches the request recorder,
- records the matched route pattern (`c.FullPath()`), Gin errors (`c.Error`) and panics. A panic is recorded and re-raised, so `gin.Recovery()` still answers.

The response writer passes `Flush` and `Hijack` through, so server-sent events and WebSockets keep working. A flushed HTML response is streamed without the toolbar.

## Testing

```bash
go test -race ./adapter/gin/...
```
