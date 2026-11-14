# gin adapter

`adapter/gin` integrates `godebugbar` with `github.com/gin-gonic/gin`.

## Usage

Pass `ginadapter.Attach(engine)` to `godebugbar.New`:

```go
engine := gin.New()

godebugbar.New(
	ginadapter.Attach(engine),
)
```

`Attach` enables `engine.ContextWithFallback = true` and registers the profiler middleware on `engine`. Matched route patterns (`ginContext.FullPath()`), status codes, response sizes, and Gin errors are recorded into each request's profile, and `/_debugbar/*` endpoints are served automatically.

## Testing

```bash
go test -race ./adapter/gin/...
```
