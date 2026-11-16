package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

func commentProblems(fileSet *token.FileSet, file *ast.File, path string) []string {
	allowed := map[*ast.CommentGroup]bool{}
	if isPublicAPIFile(path) {
		allowed = docComments(file)
	}

	var problems []string
	for _, group := range file.Comments {
		if allowed[group] || isExampleOutput(group, path) {
			continue
		}
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, directivePrefix) {
				continue
			}

			problems = append(problems, fmt.Sprintf("%s: comment not allowed: %s", fileSet.Position(comment.Pos()), firstLine(comment.Text)))
		}
	}

	return problems
}

func isExampleOutput(group *ast.CommentGroup, path string) bool {
	if !strings.HasSuffix(path, testFileSuffix) {
		return false
	}
	first := strings.TrimSpace(strings.TrimPrefix(group.List[0].Text, "//"))

	return slices.ContainsFunc(exampleOutputMarkers, func(marker string) bool {
		return strings.HasPrefix(first, marker)
	})
}

func isPublicAPIFile(path string) bool {
	if strings.HasSuffix(path, testFileSuffix) {
		return false
	}
	parts := strings.Split(filepath.ToSlash(filepath.Dir(path)), "/")

	return !slices.Contains(parts, internalDir) && !slices.Contains(parts, toolsDir)
}

func docComments(file *ast.File) map[*ast.CommentGroup]bool {
	allowed := make(map[*ast.CommentGroup]bool)
	allow := func(group *ast.CommentGroup, exported bool) {
		if group != nil && exported {
			allowed[group] = true
		}
	}

	allow(file.Doc, true)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			allow(d.Doc, d.Name.IsExported() && exportedReceiver(d))
		case *ast.GenDecl:
			allow(d.Doc, anyExportedSpec(d))
			for _, spec := range d.Specs {
				allowSpec(spec, allow)
			}
		}
	}

	return allowed
}

func allowSpec(spec ast.Spec, allow func(*ast.CommentGroup, bool)) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		allow(s.Doc, s.Name.IsExported())
		if st, ok := s.Type.(*ast.StructType); ok && s.Name.IsExported() {
			for _, field := range st.Fields.List {
				allow(field.Doc, anyExportedName(field.Names))
			}
		}
	case *ast.ValueSpec:
		allow(s.Doc, anyExportedName(s.Names))
	}
}

func anyExportedSpec(decl *ast.GenDecl) bool {
	return slices.ContainsFunc(decl.Specs, func(spec ast.Spec) bool {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			return s.Name.IsExported()
		case *ast.ValueSpec:
			return anyExportedName(s.Names)
		default:
			return false
		}
	})
}

func anyExportedName(names []*ast.Ident) bool {
	return slices.ContainsFunc(names, func(name *ast.Ident) bool { return name.IsExported() })
}

func exportedReceiver(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return true
	}

	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if index, ok := expr.(*ast.IndexExpr); ok {
		expr = index.X
	}
	ident, ok := expr.(*ast.Ident)

	return ok && ident.IsExported()
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")

	return line
}
