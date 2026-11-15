package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

func checkTree(root string) ([]string, error) {
	var problems []string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return skipDirectory(entry.Name())
		}

		if !strings.HasSuffix(path, goExtension) {
			return nil
		}

		found, err := checkFile(path)
		if err != nil {
			return err
		}

		problems = append(problems, found...)

		return nil
	})

	return problems, err
}

func skipDirectory(name string) error {
	if skippedDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
		return filepath.SkipDir
	}

	return nil
}

func checkFile(path string) ([]string, error) {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if ast.IsGenerated(file) {
		return nil, nil
	}

	problems := commentProblems(fileSet, file, path)
	problems = append(problems, receiverProblems(fileSet, file)...)
	problems = append(problems, typeProblems(file, path)...)

	return problems, nil
}
