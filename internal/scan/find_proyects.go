package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

type FindProjectsOpts struct {
	Path     string
	MaxDepth int
}

type Config struct {
	ignoredDirs  []string
	maxLoopDepth int
}

const defaultMaxLoopDepth = 100

func parseSingleFile(filePath string) ([]Project, error) {
	absFilePath := absolutePath(filePath)

	fileData, err := os.ReadFile(absFilePath)
	if err != nil {
		return nil, err
	}

	file, err := modfile.Parse(absFilePath, fileData, nil)
	if err != nil {
		return nil, err
	}

	if file.Module.Mod.Path == "" {
		return []Project{}, nil
	}

	return []Project{
		{
			Path:    filepath.Dir(absFilePath),
			Package: file.Module.Mod.Path,
		},
	}, nil

}

func parseWorkspaceEntries(workspacePath string, basePath string) ([]Project, error) {
	absWorkspacePath := absolutePath(workspacePath)
	absBasePath := absolutePath(basePath)

	fileData, err := os.ReadFile(absWorkspacePath)
	if err != nil {
		return nil, err
	}

	file, err := modfile.ParseWork(absWorkspacePath, fileData, nil)
	if err != nil {
		return nil, err
	}

	var projects []Project

	if file.Use == nil {
		return projects, nil
	}

	for _, directive := range file.Use {
		if dirPath, ok := strings.CutPrefix(directive.Path, "./"); ok {
			modPath := filepath.Join(absBasePath, dirPath, "go.mod")
			if _, err := os.Stat(modPath); err == nil {
				mods, parseErr := parseSingleFile(modPath)
				if parseErr == nil && len(mods) > 0 {
					projects = append(projects, mods...)
				}
			}
		}
	}

	return projects, nil
}

func FindProjects(opts *FindProjectsOpts) ([]Project, error) {
	if opts == nil {
		opts = &FindProjectsOpts{}
	}

	if opts.Path == "" {
		opts.Path = "."
	}

	absPath := absolutePath(opts.Path)
	if absPath == "" {
		return []Project{}, nil
	}

	visited := make(map[string]bool)

	countMod, countWork := getMarkerFileCount(absPath)

	switch {
	case countMod == 1 && countWork == 0:
		projects, err := parseSingleFile(filepath.Join(absPath, "go.mod"))
		if err == nil && len(projects) > 0 {
			return projects, nil
		}
		return []Project{}, nil
	case countMod > 1 && countWork > 0:
		projects, err := parseWorkspaceEntries(filepath.Join(absPath, "go.work"), absPath)
		if err == nil {
			return projects, nil
		}
		return []Project{}, nil
	}

	var projects []Project

	err := filepath.WalkDir(absPath, func(currentPath string, info fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		absCurrentPath := absolutePath(currentPath)
		if visited[absCurrentPath] {
			return filepath.SkipDir
		}
		visited[absCurrentPath] = true

		if info.IsDir() {
			if slices.Contains(ignoredDirs, info.Name()) {
				return filepath.SkipDir
			}

			fileName := filepath.Base(currentPath)
			if fileName == "go.mod" || fileName == "go.work" {
				modProjects, parseErr := parseSingleFile(currentPath)
				if parseErr == nil && len(modProjects) > 0 {
					projects = append(projects, modProjects...)
				}
				return filepath.SkipDir
			}

			return nil
		}

		base := filepath.Base(currentPath)
		if base == "go.mod" || base == "go.work" {
			modProjects, parseErr := parseSingleFile(currentPath)
			if parseErr == nil && len(modProjects) > 0 {
				projects = append(projects, modProjects...)
			}
			return nil
		}

		return nil
	})

	if err != nil {
		return []Project{}, nil
	}

	return projects, nil
}
