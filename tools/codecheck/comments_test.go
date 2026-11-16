package main

import "testing"

func TestCommentProblems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "no comments", path: "p.go", source: "package p\n\nvar X = 1\n"},
		{name: "compiler directive is allowed", path: "internal/p/p.go", source: "package p\n\n//go:generate echo hi\nvar X = 1\n"},
		{name: "a URL in a string is not a comment", path: "p.go", source: "package p\n\nvar X = \"https://example.com\"\n"},
		{name: "doc comment on an exported var", path: "p.go", source: "package p\n\n// X is one.\nvar X = 1\n"},
		{name: "package doc comment", path: "adapter/gin/doc.go", source: "// Package p does things.\npackage p\n"},
		{name: "doc comment on an exported func", path: "p.go", source: "package p\n\n// F does it.\nfunc F() {}\n"},
		{name: "doc comment on an exported method", path: "p.go", source: "package p\n\ntype T struct{}\n\n// M does it.\nfunc (t *T) M() {}\n"},
		{name: "doc comment on an exported field", path: "p.go", source: "package p\n\ntype T struct {\n\t// A is a.\n\tA int\n}\n"},
		{name: "doc comment on an unexported var", path: "p.go", source: "package p\n\n// x is one\nvar x = 1\n", want: []string{"comment not allowed: // x is one"}},
		{name: "doc comment on a method of an unexported type", path: "p.go", source: "package p\n\ntype t struct{}\n\n// M does it.\nfunc (v *t) M() {}\n", want: []string{"comment not allowed: // M does it."}},
		{name: "comment inside a function", path: "p.go", source: "package p\n\nfunc F() {\n\t// step\n}\n", want: []string{"comment not allowed: // step"}},
		{name: "inline comment", path: "p.go", source: "package p\n\nvar X = 1 // one\n", want: []string{"comment not allowed"}},
		{name: "doc comment in internal code", path: "internal/engine/e.go", source: "package p\n\n// X is one.\nvar X = 1\n", want: []string{"comment not allowed: // X is one."}},
		{name: "doc comment in a test file", path: "p_test.go", source: "package p\n\n// X is one.\nvar X = 1\n", want: []string{"comment not allowed"}},
		{name: "example output in a test file", path: "example_test.go", source: "package p\n\nfunc ExampleF() {\n\t// Output:\n\t// 42\n}\n"},
		{name: "one-line example output in a test file", path: "example_test.go", source: "package p\n\nfunc ExampleF() {\n\t// Output: 42\n}\n"},
		{name: "unordered example output in a test file", path: "example_test.go", source: "package p\n\nfunc ExampleF() {\n\t// Unordered output:\n\t// a\n}\n"},
		{name: "example output outside a test file", path: "p.go", source: "package p\n\nfunc F() {\n\t// Output:\n\t// 42\n}\n", want: []string{"comment not allowed: // Output:", "comment not allowed: // 42"}},
		{name: "doc comment in tools", path: "tools/codecheck/main.go", source: "package p\n\n// X is one.\nvar X = 1\n", want: []string{"comment not allowed"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fileSet, file := parse(t, tt.source)

			assertProblems(t, commentProblems(fileSet, file, tt.path), tt.want)
		})
	}
}
