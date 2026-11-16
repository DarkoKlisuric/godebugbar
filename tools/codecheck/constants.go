package main

const (
	goExtension     = ".go"
	testFileSuffix  = "_test.go"
	directivePrefix = "//go:"
	readmeFile      = "README.md"
	internalDir     = "internal"
	toolsDir        = "tools"

	exitProblems = 1
	exitError    = 2
)

var skippedDirs = map[string]bool{"vendor": true, "node_modules": true}

var exampleOutputMarkers = []string{"Output:", "Unordered output:"}
