package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, content := range files {
		writeFile(t, filepath.Join(root, name), content)
	}

	return root
}

func parse(t *testing.T, source string) (*token.FileSet, *ast.File) {
	t.Helper()

	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, "file.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	return fileSet, file
}

func assertProblems(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("problems = %v, want %d problem(s)", got, len(want))
	}

	for i, part := range want {
		if !strings.Contains(got[i], part) {
			t.Errorf("problem %q does not mention %q", got[i], part)
		}
	}
}
