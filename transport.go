package godebugbar

import (
	"net/http"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
)

// Transport is an http.RoundTripper that records outgoing calls made with a
// profiled request's context.
type Transport struct {
	next http.RoundTripper
}

// NewTransport wraps next, or http.DefaultTransport when next is nil.
func NewTransport(next http.RoundTripper) *Transport {
	if next == nil {
		next = http.DefaultTransport
	}

	return &Transport{next: next}
}

// RoundTrip performs the call and records it, with credentials in the URL redacted.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := contextstate.RecorderFromContext(req.Context())
	if rec == nil {
		return t.next.RoundTrip(req)
	}

	start := time.Now()
	res, err := t.next.RoundTrip(req)
	entry := contextstate.HTTPCallEntry{
		Method:    req.Method,
		URL:       req.URL.String(),
		Duration:  time.Since(start),
		StartedAt: start,
	}
	if res != nil {
		entry.Status = res.StatusCode
	}
	if err != nil {
		entry.Error = err.Error()
	}
	rec.AddHTTPCall(entry)

	return res, err
}
