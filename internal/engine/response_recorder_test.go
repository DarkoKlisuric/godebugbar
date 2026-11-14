package engine_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestResponseRecorderHTMLInjection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		contentType string
		body        string
		snippet     string
		wantSubstr  string
	}{
		{
			name:        "injects before closing body tag",
			contentType: "text/html; charset=utf-8",
			body:        "<html><body><h1>Hello</h1></body></html>",
			snippet:     "<script src=\"/_debugbar/assets/toolbar.js\"></script>",
			wantSubstr:  "<h1>Hello</h1><script src=\"/_debugbar/assets/toolbar.js\"></script></body>",
		},
		{
			name:        "appends to html without body tag",
			contentType: "text/html",
			body:        "<div>fragment</div>",
			snippet:     "<script></script>",
			wantSubstr:  "<div>fragment</div><script></script>",
		},
		{
			name:        "skips injection for json",
			contentType: "application/json",
			body:        `{"ok":true}`,
			snippet:     "<script></script>",
			wantSubstr:  `{"ok":true}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			rr := engine.NewResponseRecorder(rec)
			if tc.contentType != "" {
				rr.Header().Set("Content-Type", tc.contentType)
			}
			rr.WriteHeader(http.StatusCreated)
			_, _ = rr.Write([]byte(tc.body))
			rr.FlushWithSnippet(tc.snippet)

			got := rec.Body.String()
			if !strings.Contains(got, tc.wantSubstr) {
				t.Fatalf("expected body %q to contain %q", got, tc.wantSubstr)
			}
			if tc.contentType == "application/json" && strings.Contains(got, tc.snippet) {
				t.Fatalf("expected json response not to contain snippet")
			}
			if rr.StatusCode() != http.StatusCreated {
				t.Fatalf("expected status 201, got %d", rr.StatusCode())
			}
			if rr.Size() != len(tc.body) {
				t.Fatalf("expected size %d, got %d", len(tc.body), rr.Size())
			}
		})
	}
}
