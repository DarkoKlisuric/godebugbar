package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func commentProblems(fileSet *token.FileSet, file *ast.File) []string {
	var problems []string
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, directivePrefix) {
				continue
			}

			problems = append(problems, fmt.Sprintf("%s: comment not allowed: %s", fileSet.Position(comment.Pos()), firstLine(comment.Text)))
		}
	}

	return problems
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")

	return line
}
