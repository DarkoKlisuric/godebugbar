package ginadapter

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

type ginAdapter struct {
	eng *engine.Engine
}

func Attach(router *gin.Engine) godebugbar.Option {
	return func(p *godebugbar.Profiler) {
		if router == nil || p == nil {
			return
		}
		eng := engine.Lookup(p)
		if eng == nil {
			return
		}
		eng.AddCollectors(collector.NewRequestCollector())
		eng.AddInitializer(func(e *engine.Engine) {
			a := &ginAdapter{eng: e}
			a.mount(router)
			router.Use(a.middleware())
		})
	}
}

func (a *ginAdapter) mount(router *gin.Engine) {
	if !a.eng.Enabled() {
		return
	}
	handler := gin.WrapH(a.eng.Handler())
	prefix := a.eng.Prefix()
	if a.eng.DashboardEnabled() {
		router.GET(prefix, handler)
		router.GET(prefix+"/", handler)
	}
	router.GET(prefix+"/assets/toolbar.js", handler)
	router.GET(prefix+"/api/profiles", handler)
	router.GET(prefix+"/api/profiles/latest", handler)
	router.GET(prefix+"/api/profiles/:id", handler)
	router.OPTIONS(prefix+"/api/profiles", handler)
	router.OPTIONS(prefix+"/api/profiles/latest", handler)
	router.OPTIONS(prefix+"/api/profiles/:id", handler)
}

func (a *ginAdapter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.eng.Enabled() {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		prefix := a.eng.Prefix()
		if path == prefix || strings.HasPrefix(path, prefix+"/") || a.eng.ShouldSkip(c.Request.Method, path) {
			c.Next()
			return
		}

		rec := contextstate.NewRecorder()
		ctx := contextstate.WithRecorder(c.Request.Context(), rec)
		c.Request = c.Request.WithContext(ctx)

		gw := &ginResponseWriter{
			ResponseWriter: c.Writer,
			allowHTML:      a.eng.HTMLInjectionEnabled(),
		}
		c.Writer = gw

		profileID := a.eng.NewProfileID()
		c.Header(engine.ProfileHeader, profileID)
		a.eng.ExposeProfileHeader(c.Writer.Header())

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		if route := c.FullPath(); route != "" {
			rec.SetRoute(route)
		}

		for _, ginErr := range c.Errors {
			if ginErr != nil && ginErr.Err != nil {
				rec.AddError(ginErr.Err.Error())
			}
		}

		a.eng.RecordRequest(c.Request, profileID, start, duration, &contract.ResponseInfo{
			Status:   gw.Status(),
			Size:     gw.size,
			Duration: duration,
			Headers:  gw.Header().Clone(),
			Route:    rec.Route(),
			Errors:   rec.Errors(),
		})

		snippet := ""
		if a.eng.HTMLInjectionEnabled() {
			snippet = a.eng.Snippet(profileID)
		}
		gw.flushWithSnippet(snippet)
	}
}

type ginResponseWriter struct {
	gin.ResponseWriter
	buf       bytes.Buffer
	allowHTML bool
	isHTML    bool
	decided   bool
	size      int
}

func (w *ginResponseWriter) Write(b []byte) (int, error) {
	w.decide(b)
	w.size += len(b)
	if w.isHTML {
		return w.buf.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *ginResponseWriter) WriteString(s string) (int, error) {
	w.decide([]byte(s))
	w.size += len(s)
	if w.isHTML {
		return w.buf.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *ginResponseWriter) decide(sample []byte) {
	if w.decided {
		return
	}
	w.decided = true
	if !w.allowHTML {
		return
	}
	ct := w.Header().Get("Content-Type")
	if ct == "" && len(sample) > 0 {
		ct = "text/plain"
		if strings.Contains(strings.ToLower(string(sample)), "<html") || strings.Contains(strings.ToLower(string(sample)), "<body") {
			ct = "text/html; charset=utf-8"
		}
	}
	enc := w.Header().Get("Content-Encoding")
	w.isHTML = strings.Contains(strings.ToLower(ct), "text/html") && enc == ""
}

func (w *ginResponseWriter) flushWithSnippet(snippet string) {
	if !w.isHTML {
		return
	}
	body := w.buf.String()
	lower := strings.ToLower(body)
	idx := strings.LastIndex(lower, "</body>")
	var out string
	if idx != -1 {
		out = body[:idx] + snippet + body[idx:]
	} else {
		out = body + snippet
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(out)))
	_, _ = w.ResponseWriter.Write([]byte(out))
}
