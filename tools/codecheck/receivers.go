package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func receiverProblems(fileSet *token.FileSet, file *ast.File) []string {
	var problems []string
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Recv == nil || len(function.Recv.List) == 0 {
			continue
		}

		if _, isPointer := function.Recv.List[0].Type.(*ast.StarExpr); isPointer {
			continue
		}

		problems = append(problems, fmt.Sprintf("%s: method %s needs a pointer receiver", fileSet.Position(function.Pos()), function.Name.Name))
	}

	return problems
}
