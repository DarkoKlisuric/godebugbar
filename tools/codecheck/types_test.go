package main

import "testing"

func TestTypeProblems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		source   string
		want     []string
	}{
		{name: "one type named after its file", filename: "query_collector.go", source: "package p\n\ntype QueryCollector struct{}\n"},
		{name: "acronyms", filename: "sql_collector.go", source: "package p\n\ntype SQLCollector struct{}\n"},
		{name: "test helper types live with their tests", filename: "profiler_test.go", source: "package p\n\ntype fakeCollector struct{}\n"},
		{name: "no types", filename: "constants.go", source: "package p\n\nconst A = 1\n"},
		{name: "aliases do not count", filename: "alias.go", source: "package p\n\ntype (\n\tA = int\n\tB = string\n)\n"},
		{name: "small types next to the main one", filename: "collector.go", source: "package p\n\ntype BadgeColor string\n\ntype Collector interface{}\n"},
		{
			name:     "several types, none named after the file",
			filename: "types.go",
			source:   "package p\n\ntype Profile struct{}\n\ntype Section struct{}\n",
			want:     []string{"no type named after the file; name the file after its main type (profile.go)"},
		},
		{
			name:     "type in a file with another name",
			filename: "profiles.go",
			source:   "package p\n\ntype Profile struct{}\n",
			want:     []string{"no type named after the file; name the file after its main type (profile.go)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, file := parse(t, tt.source)

			assertProblems(t, typeProblems(file, tt.filename), tt.want)
		})
	}
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "one word", in: "Profile", want: "profile"},
		{name: "two words", in: "QueryCollector", want: "query_collector"},
		{name: "unexported", in: "responseRecorder", want: "response_recorder"},
		{name: "acronym at the end", in: "RequestID", want: "request_id"},
		{name: "acronym at the start", in: "SQLCollector", want: "sql_collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := snakeCase(tt.in); got != tt.want {
				t.Errorf("snakeCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
