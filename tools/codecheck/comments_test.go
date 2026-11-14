package main

import "testing"

func TestCommentProblems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{name: "no comments", source: "package p\n\nvar X = 1\n"},
		{name: "compiler directive is allowed", source: "package p\n\n//go:generate echo hi\nvar X = 1\n"},
		{name: "line comment", source: "package p\n\n// X is one\nvar X = 1\n", want: []string{"comment not allowed: // X is one"}},
		{name: "inline comment", source: "package p\n\nvar X = 1 // one\n", want: []string{"comment not allowed"}},
		{name: "a URL in a string is not a comment", source: "package p\n\nvar X = \"https://example.com\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fileSet, file := parse(t, tt.source)

			assertProblems(t, commentProblems(fileSet, file), tt.want)
		})
	}
}
