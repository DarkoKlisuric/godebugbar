package engine_test

import (
	"io"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestBodyCapture(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		body          string
		limit         int
		wantCaptured  string
		wantTruncated bool
	}{
		{name: "fits", body: "hello", limit: 10, wantCaptured: "hello"},
		{name: "cut at the limit", body: "hello world", limit: 5, wantCaptured: "hello", wantTruncated: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			capture := engine.NewBodyCapture(io.NopCloser(strings.NewReader(tc.body)), tc.limit)
			read, err := io.ReadAll(capture)
			if err != nil || string(read) != tc.body {
				t.Fatalf("handler must still read the whole body, got %q %v", read, err)
			}
			captured, truncated := capture.Captured()
			if string(captured) != tc.wantCaptured || truncated != tc.wantTruncated {
				t.Fatalf("Captured() = %q %v, want %q %v", captured, truncated, tc.wantCaptured, tc.wantTruncated)
			}
			if err := capture.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
		})
	}
}
