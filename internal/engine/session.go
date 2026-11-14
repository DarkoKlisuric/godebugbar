package engine

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/redact"
)

const (
	truncatedBodySuffix     = "\n… [truncated]"
	truncatedStructuredBody = "[body larger than %d bytes is not shown: it cannot be redacted when cut off]"
)

type Session struct {
	eng        *Engine
	Recorder   *contextstate.Recorder
	Request    *http.Request
	ProfileID  string
	Start      time.Time
	authorized bool
	body       *BodyCapture
}

func (e *Engine) Begin(r *http.Request) *Session {
	recorder := contextstate.NewRecorderWithConfig(e.recorderConfig)
	req := r.WithContext(contextstate.WithRecorder(r.Context(), recorder))

	s := &Session{
		eng:        e,
		Recorder:   recorder,
		ProfileID:  e.NewProfileID(),
		Start:      time.Now(),
		authorized: e.Authorized(r),
	}
	if e.requestBodyLimit > 0 && r.Body != nil && r.Body != http.NoBody {
		s.body = NewBodyCapture(r.Body, e.requestBodyLimit)
		req.Body = s.body
	}
	s.Request = req

	return s
}

func (s *Session) PrepareHeaders(h http.Header) {
	if !s.authorized {
		return
	}
	h.Set(ProfileHeader, s.ProfileID)
	s.eng.ExposeProfileHeader(h)
}

func (s *Session) Snippet() string {
	if !s.authorized || !s.eng.htmlInjection {
		return ""
	}

	return s.eng.Snippet(s.ProfileID)
}

func (s *Session) RecordPanic(v any) {
	s.Recorder.SetPanic(fmt.Sprint(v), string(debug.Stack()))
}

func (s *Session) End(res *contract.ResponseInfo) Profile {
	duration := time.Since(s.Start)
	if s.body != nil {
		s.captureBody()
	}
	res.Duration = duration
	res.Route = s.Recorder.Route()
	res.Errors = s.Recorder.Errors()
	res.Panic, res.PanicStack = s.Recorder.Panic()

	return s.eng.RecordRequest(s.Request, s.ProfileID, s.Start, duration, res)
}

func (s *Session) captureBody() {
	raw, truncated := s.body.Captured()
	if len(raw) == 0 {
		return
	}

	contentType := s.Request.Header.Get("Content-Type")
	switch {
	case !truncated:
		s.Recorder.SetRequestBody(s.eng.redactor.Body(contentType, raw))
	case redact.IsStructured(contentType):
		s.Recorder.SetRequestBody(fmt.Sprintf(truncatedStructuredBody, len(raw)))
	default:
		s.Recorder.SetRequestBody(string(raw) + truncatedBodySuffix)
	}
}
