package ginadapter

import (
	"bytes"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
)

const (
	contentTypeHeader  = "Content-Type"
	contentLenHeader   = "Content-Length"
	exposeHeaderName   = "Access-Control-Expose-Headers"
	htmlContentPrefix  = "text/html"
	bodyCloseTag       = "</body>"
	unmatchedRoutePath = "unmatched"
)

type GinAdapter struct {
	profiler *godebugbar.Profiler
}

func NewGinAdapter(profiler *godebugbar.Profiler) *GinAdapter {
	return &GinAdapter{profiler: profiler}
}

func Attach(engine *gin.Engine) godebugbar.Option {
	return func(profiler *godebugbar.Profiler) {
		if engine == nil {
			return
		}

		engine.ContextWithFallback = true
		adapter := NewGinAdapter(profiler)
		engine.Use(adapter.Middleware())
	}
}

func Middleware(profiler *godebugbar.Profiler) gin.HandlerFunc {
	return NewGinAdapter(profiler).Middleware()
}

func (a *GinAdapter) Middleware() gin.HandlerFunc {
	debugHandler := a.profiler.Handler()

	return func(ginContext *gin.Context) {
		if !a.profiler.Enabled() {
			ginContext.Next()

			return
		}

		path := ginContext.Request.URL.Path
		prefix := a.profiler.Prefix()
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			debugHandler.ServeHTTP(ginContext.Writer, ginContext.Request)
			ginContext.Abort()

			return
		}

		if a.profiler.IsExcluded(path) {
			ginContext.Next()

			return
		}

		profileID := a.profiler.NewProfileID()
		startedAt := time.Now().UTC()
		recorder := godebugbar.NewRecorder()
		ctx := godebugbar.WithRecorder(ginContext.Request.Context(), recorder)
		ginContext.Request = ginContext.Request.WithContext(ctx)

		ginContext.Header(godebugbar.ProfileHeader, profileID)
		ginContext.Writer.Header().Add(exposeHeaderName, godebugbar.ProfileHeader)

		writer := &ginBodyWriter{
			ResponseWriter: ginContext.Writer,
		}
		ginContext.Writer = writer

		ginContext.Next()

		route := recorder.Route()
		if route == "" {
			route = ginContext.FullPath()
		}
		if route == "" {
			route = unmatchedRoutePath
		}

		for _, ginErr := range ginContext.Errors {
			if ginErr != nil && ginErr.Err != nil {
				recorder.AddError(ginErr.Err.Error())
			}
		}

		responseInfo := &godebugbar.ResponseInfo{
			Status:   ginContext.Writer.Status(),
			Size:     ginContext.Writer.Size(),
			Duration: time.Since(startedAt),
			Headers:  ginContext.Writer.Header().Clone(),
			Route:    route,
			Errors:   recorder.Errors(),
		}

		a.profiler.RecordRequest(ctx, profileID, startedAt, ginContext.Request, responseInfo)
		writer.flushHTML(a.profiler.Snippet(profileID))
	}
}

type ginBodyWriter struct {
	gin.ResponseWriter
	buffered bool
	body     bytes.Buffer
}

func (w *ginBodyWriter) Write(data []byte) (int, error) {
	if !w.buffered && strings.HasPrefix(strings.ToLower(w.Header().Get(contentTypeHeader)), htmlContentPrefix) {
		w.buffered = true
	}

	if w.buffered {
		return w.body.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *ginBodyWriter) WriteString(str string) (int, error) {
	return w.Write([]byte(str))
}

func (w *ginBodyWriter) flushHTML(snippet string) {
	if !w.buffered {
		return
	}

	content := w.body.String()
	if snippet != "" {
		idx := strings.LastIndex(strings.ToLower(content), bodyCloseTag)
		if idx >= 0 {
			content = content[:idx] + snippet + content[idx:]
		} else {
			content += snippet
		}
	}

	w.Header().Del(contentLenHeader)
	_, _ = w.ResponseWriter.Write([]byte(content))
}
