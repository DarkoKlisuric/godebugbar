package engine_test

import (
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestInjectSnippet(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		body    string
		snippet string
		want    string
	}{
		{name: "before the last closing body tag", body: "<body></body><!-- </body> -->", snippet: "S", want: "<body></body><!-- S</body> -->"},
		{name: "mixed case tag", body: "<Body>x</BoDy>", snippet: "S", want: "<Body>xS</BoDy>"},
		{name: "no tag appends", body: "<p>x</p>", snippet: "S", want: "<p>x</p>S"},
		{name: "empty snippet keeps body", body: "<body></body>", snippet: "", want: "<body></body>"},
		{name: "multi-byte text keeps byte offsets", body: "ßİ</body>", snippet: "S", want: "ßİS</body>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := string(engine.InjectSnippet([]byte(tc.body), tc.snippet)); got != tc.want {
				t.Fatalf("InjectSnippet() = %q, want %q", got, tc.want)
			}
		})
	}
}
