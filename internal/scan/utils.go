package scan

import (
	"io/fs"
	"path/filepath"
)

var ignoredDirs = []string{
	".git",
	"vendor",
	"node_modules",
	".cache",
	".idea",
	"bin",
	"dist",
	"build",
	".tmp",
	".opencode",
	".vscode",
}

func absolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func getMarkerFileCount(path string) (int, int) {
	var countMod, countWork int

	err := filepath.WalkDir(path, func(currentPath string, info fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		base := filepath.Base(currentPath)
		switch base {
		case "go.mod":
			countMod++
		case "go.work":
			countWork++
		}

		return nil
	})

	if err != nil {
		return 0, 0
	}

	return countMod, countWork
}
