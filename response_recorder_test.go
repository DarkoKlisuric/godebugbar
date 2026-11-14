package godebugbar

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseRecorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		body        string
		snippet     string
		wantSubstr  string
	}{
		{
			name:        "injects snippet before closing body tag on html response",
			contentType: "text/html; charset=utf-8",
			body:        "<html><body><h1>Hello</h1></body></html>",
			snippet:     "<script src=\"/_debugbar/toolbar.js\"></script>",
			wantSubstr:  "<h1>Hello</h1><script src=\"/_debugbar/toolbar.js\"></script></body>",
		},
		{
			name:        "does not modify json response",
			contentType: "application/json",
			body:        `{"ok":true}`,
			snippet:     "<script></script>",
			wantSubstr:  `{"ok":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			wrapped := newResponseRecorder(rec)
			wrapped.Header().Set("Content-Type", tt.contentType)
			wrapped.WriteHeader(http.StatusCreated)
			_, _ = wrapped.Write([]byte(tt.body))
			wrapped.FlushBuffered(tt.snippet)

			if rec.Code != http.StatusCreated {
				t.Errorf("Code = %d, want %d", rec.Code, http.StatusCreated)
			}

			if !strings.Contains(rec.Body.String(), tt.wantSubstr) {
				t.Errorf("Body = %q, want substring %q", rec.Body.String(), tt.wantSubstr)
			}
		})
	}
}
