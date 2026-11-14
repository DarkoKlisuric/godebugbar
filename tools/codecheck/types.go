package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"unicode"
)

func typeProblems(file *ast.File, path string) []string {
	name := filepath.Base(path)
	if strings.HasSuffix(name, testFileSuffix) {
		return nil
	}

	types := declaredTypes(file)
	if len(types) == 0 {
		return nil
	}

	for _, typeName := range types {
		if fileStem(name) == snakeCase(typeName) {
			return nil
		}
	}

	return []string{fmt.Sprintf("%s: no type named after the file; name the file after its main type (%s)", path, expectedFileName(types[0], name))}
}

func declaredTypes(file *ast.File) []string {
	var names []string
	for _, spec := range typeSpecs(file) {
		if !spec.Assign.IsValid() {
			names = append(names, spec.Name.Name)
		}
	}

	return names
}

func typeSpecs(file *ast.File) []*ast.TypeSpec {
	var specs []*ast.TypeSpec
	for _, decl := range file.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}

		for _, spec := range general.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				specs = append(specs, typeSpec)
			}
		}
	}

	return specs
}

func fileStem(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, testFileSuffix), goExtension)
}

func expectedFileName(typeName, actual string) string {
	if strings.HasSuffix(actual, testFileSuffix) {
		return snakeCase(typeName) + testFileSuffix
	}

	return snakeCase(typeName) + goExtension
}

func snakeCase(name string) string {
	runes := []rune(name)

	var builder strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && startsWord(runes, i) {
			builder.WriteRune('_')
		}

		builder.WriteRune(unicode.ToLower(r))
	}

	return builder.String()
}

func startsWord(runes []rune, i int) bool {
	previousIsLower := unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])
	nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

	return previousIsLower || nextIsLower
}
