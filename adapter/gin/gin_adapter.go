package ginadapter

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

const routesRegisteredMessage = "godebugbar: ginadapter.Attach must run before any route is registered; %d routes would not be profiled"

type ginAdapter struct {
	eng *engine.Engine
}

func Attach(router *gin.Engine) godebugbar.Option {
	return func(e *engine.Engine) {
		if router == nil {
			return
		}
		e.AddInitializer(func(e *engine.Engine) {
			if !e.Enabled() {
				return
			}
			if routes := router.Routes(); len(routes) > 0 {
				panic(fmt.Sprintf(routesRegisteredMessage, len(routes)))
			}
			router.ContextWithFallback = true
			a := &ginAdapter{eng: e}
			router.Use(a.middleware())
			a.mount(router)
		})
	}
}

func (a *ginAdapter) mount(router *gin.Engine) {
	if !a.eng.ServesEndpoints() {
		return
	}
	handler := gin.WrapH(a.eng.Handler())
	prefix := a.eng.Prefix()
	router.Any(prefix, handler)
	router.Any(prefix+"/*path", handler)
}

func (a *ginAdapter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if a.eng.IsEndpoint(path) || a.eng.ShouldSkip(c.Request.Method, path) {
			c.Next()
			return
		}

		session := a.eng.Begin(c.Request)
		c.Request = session.Request

		gw := &ginResponseWriter{
			ResponseWriter: c.Writer,
			allowHTML:      a.eng.HTMLInjectionEnabled(),
			headRequest:    c.Request.Method == http.MethodHead,
			maxBuffer:      a.eng.MaxBuffer(),
		}
		c.Writer = gw
		session.PrepareHeaders(gw.Header())

		defer func() {
			if v := recover(); v != nil {
				session.RecordPanic(v)
				a.finish(c, session, gw, http.StatusInternalServerError)
				panic(v)
			}
		}()
		c.Next()

		a.finish(c, session, gw, gw.Status())
		gw.finish(session.Snippet())
	}
}

func (a *ginAdapter) finish(c *gin.Context, session *engine.Session, gw *ginResponseWriter, status int) {
	if route := c.FullPath(); route != "" {
		session.Recorder.SetRoute(route)
	}
	for _, ginErr := range c.Errors {
		if ginErr != nil && ginErr.Err != nil {
			session.Recorder.AddError(ginErr.Err.Error())
		}
	}

	session.End(&contract.ResponseInfo{
		Status:  status,
		Size:    gw.size,
		Headers: gw.Header().Clone(),
	})
}

type ginResponseWriter struct {
	gin.ResponseWriter
	buf         bytes.Buffer
	allowHTML   bool
	headRequest bool
	maxBuffer   int
	buffering   bool
	decided     bool
	hijacked    bool
	size        int
}

func (w *ginResponseWriter) Write(b []byte) (int, error) {
	w.decide(b)
	w.size += len(b)
	if w.buffering && w.buf.Len()+len(b) > w.maxBuffer {
		if err := w.passthrough(); err != nil {
			return 0, err
		}
	}
	if w.buffering {
		return w.buf.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func (w *ginResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *ginResponseWriter) Flush() {
	_ = w.passthrough()
	w.ResponseWriter.Flush()
}

func (w *ginResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := w.ResponseWriter.Hijack()
	if err == nil {
		w.hijacked = true
	}

	return conn, rw, err
}

func (w *ginResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *ginResponseWriter) decide(sample []byte) {
	if w.decided {
		return
	}
	w.decided = true
	if !w.allowHTML || w.headRequest || w.Written() {
		return
	}
	status := w.Status()
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return
	}
	header := w.Header()
	if header.Get("Content-Type") == "" && len(sample) > 0 {
		header.Set("Content-Type", http.DetectContentType(sample))
	}
	w.buffering = engine.IsInjectableHTML(header)
}

func (w *ginResponseWriter) passthrough() error {
	if !w.buffering {
		return nil
	}
	w.buffering = false
	_, err := w.ResponseWriter.Write(w.buf.Bytes())
	w.buf.Reset()

	return err
}

func (w *ginResponseWriter) finish(snippet string) {
	if w.hijacked || !w.buffering {
		return
	}
	out := engine.InjectSnippet(w.buf.Bytes(), snippet)
	w.buffering = false
	w.buf.Reset()
	w.Header().Set("Content-Length", strconv.Itoa(len(out)))
	_, _ = w.ResponseWriter.Write(out)
}
