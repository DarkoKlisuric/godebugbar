package main

const (
	goExtension     = ".go"
	testFileSuffix  = "_test.go"
	directivePrefix = "//go:"
	readmeFile      = "README.md"

	exitProblems = 1
	exitError    = 2
)

var skippedDirs = map[string]bool{"vendor": true, "node_modules": true}
