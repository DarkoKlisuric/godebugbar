package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func readmeProblems(root string) ([]string, error) {
	goDirs := make(map[string]bool)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return skipDirectory(entry.Name())
		}

		if strings.HasSuffix(path, goExtension) {
			goDirs[filepath.Dir(path)] = true
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(goDirs))
	for dir := range goDirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	var problems []string
	for _, dir := range dirs {
		readmePath := filepath.Join(dir, readmeFile)
		if _, statErr := os.Stat(readmePath); os.IsNotExist(statErr) {
			problems = append(problems, fmt.Sprintf("%s: missing %s", dir, readmeFile))
		}
	}

	return problems, nil
}
