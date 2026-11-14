package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	found, err := run(root, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}

	if found > 0 {
		_, _ = fmt.Printf("%d problem(s) found; see AGENTS.md\n", found)
		os.Exit(exitProblems)
	}
}

func run(root string, out io.Writer) (int, error) {
	fileProblems, err := checkTree(root)
	if err != nil {
		return 0, err
	}

	readmes, err := readmeProblems(root)
	if err != nil {
		return 0, err
	}

	problems := fileProblems
	problems = append(problems, readmes...)
	for _, problem := range problems {
		_, _ = fmt.Fprintln(out, problem)
	}

	return len(problems), nil
}
