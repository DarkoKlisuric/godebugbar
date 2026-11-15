package main

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     map[string]string
		wantFound int
		wantOut   []string
	}{
		{
			name: "clean project",
			files: map[string]string{
				"README.md":   "# godebugbar\n",
				"profiler.go": "package godebugbar\n\ntype Profiler struct{}\n",
			},
		},
		{
			name: "every problem is reported once",
			files: map[string]string{
				"internal/collector/queries.go": "package collector\n\n// QueryCollector collects queries.\ntype QueryCollector struct{}\n\nfunc (c QueryCollector) Name() string { return \"\" }\n",
			},
			wantFound: 4,
			wantOut:   []string{"missing README.md", "comment not allowed", "pointer receiver", "name the file after its main type (query_collector.go)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out strings.Builder

			found, err := run(writeFiles(t, tt.files), &out)
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}

			if found != tt.wantFound {
				t.Errorf("run() found = %d, want %d\noutput:\n%s", found, tt.wantFound, out.String())
			}

			for _, part := range tt.wantOut {
				if !strings.Contains(out.String(), part) {
					t.Errorf("output missing %q\noutput:\n%s", part, out.String())
				}
			}
		})
	}
}
