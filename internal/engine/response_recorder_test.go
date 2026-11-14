package engine_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

const testSnippet = `<script src="/_debugbar/assets/toolbar.js"></script>`

func TestResponseRecorderHTMLInjection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		method      string
		inject      bool
		maxBuffer   int
		contentType string
		status      int
		body        string
		want        string
	}{
		{
			name:        "injects before closing body tag",
			inject:      true,
			contentType: "text/html; charset=utf-8",
			status:      http.StatusCreated,
			body:        "<html><body><h1>Hello</h1></body></html>",
			want:        "<html><body><h1>Hello</h1>" + testSnippet + "</body></html>",
		},
		{
			name:        "matches an uppercase closing tag",
			inject:      true,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<HTML><BODY>x</BODY></HTML>",
			want:        "<HTML><BODY>x" + testSnippet + "</BODY></HTML>",
		},
		{
			name:        "keeps position with multi-byte text before the tag",
			inject:      true,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<body>" + strings.Repeat("İ", 21) + "<p>x</p></body>",
			want:        "<body>" + strings.Repeat("İ", 21) + "<p>x</p>" + testSnippet + "</body>",
		},
		{
			name:        "appends to html without body tag",
			inject:      true,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<div>fragment</div>",
			want:        "<div>fragment</div>" + testSnippet,
		},
		{
			name:        "skips json",
			inject:      true,
			contentType: "application/json",
			status:      http.StatusOK,
			body:        `{"ok":true}`,
			want:        `{"ok":true}`,
		},
		{
			name:        "skips html when injection is off",
			inject:      false,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<body></body>",
			want:        "<body></body>",
		},
		{
			name:        "skips head requests",
			method:      http.MethodHead,
			inject:      true,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<body></body>",
			want:        "<body></body>",
		},
		{
			name:        "streams bodies larger than the buffer cap",
			inject:      true,
			maxBuffer:   8,
			contentType: "text/html",
			status:      http.StatusOK,
			body:        "<body>large page</body>",
			want:        "<body>large page</body>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			method := tc.method
			if method == "" {
				method = http.MethodGet
			}
			req := httptest.NewRequestWithContext(t.Context(), method, "/", nil)
			rec := httptest.NewRecorder()
			rr := engine.NewResponseRecorder(rec, req, tc.inject, tc.maxBuffer)
			rr.Header().Set("Content-Type", tc.contentType)
			rr.WriteHeader(tc.status)
			_, _ = rr.Write([]byte(tc.body))
			rr.Finish(testSnippet)

			if got := rec.Body.String(); got != tc.want {
				t.Fatalf("body = %q, want %q", got, tc.want)
			}
			if rec.Code != tc.status || rr.StatusCode() != tc.status {
				t.Fatalf("status = %d/%d, want %d", rec.Code, rr.StatusCode(), tc.status)
			}
			if rr.Size() != len(tc.body) {
				t.Fatalf("size = %d, want %d", rr.Size(), len(tc.body))
			}
		})
	}
}

func TestResponseRecorderFlushStopsBuffering(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rr := engine.NewResponseRecorder(rec, req, true, 0)
	rr.Header().Set("Content-Type", "text/html")
	_, _ = rr.Write([]byte("<body>first"))

	if err := http.NewResponseController(rr).Flush(); err != nil {
		t.Fatalf("Flush through ResponseController: %v", err)
	}
	if !rec.Flushed || rec.Body.String() != "<body>first" {
		t.Fatalf("expected buffered bytes to be flushed, got flushed=%v body=%q", rec.Flushed, rec.Body.String())
	}

	_, _ = rr.Write([]byte("</body>"))
	rr.Finish(testSnippet)
	if strings.Contains(rec.Body.String(), testSnippet) {
		t.Fatalf("expected no injection after a flush, got %q", rec.Body.String())
	}
}

func TestResponseRecorderHijackUsesUnderlyingWriter(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rr := engine.NewResponseRecorder(httptest.NewRecorder(), req, false, 0)

	_, _, err := http.NewResponseController(rr).Hijack()
	if !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("expected the underlying writer's ErrNotSupported, got %v", err)
	}
	if rr.Unwrap() == nil {
		t.Fatalf("expected Unwrap to return the underlying writer")
	}
}

func TestResponseRecorderInformationalStatus(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rr := engine.NewResponseRecorder(rec, req, false, 0)
	rr.WriteHeader(http.StatusEarlyHints)
	rr.WriteHeader(http.StatusTeapot)
	rr.Finish("")

	if rr.StatusCode() != http.StatusTeapot {
		t.Fatalf("recorded status = %d, want %d", rr.StatusCode(), http.StatusTeapot)
	}
}

func TestResponseRecorderWritesStatusWithoutBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rr := engine.NewResponseRecorder(rec, req, true, 0)
	rr.Finish(testSnippet)

	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("expected empty 200, got %d %q", rec.Code, rec.Body.String())
	}
}
