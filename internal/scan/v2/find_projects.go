package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/padiazg/test-finder/internal/project"
	"github.com/padiazg/test-finder/pkg/helpers"
)

type Finder struct {
	path         string
	ignoredDirs  []string
	maxLoopDepth int
	full         bool
}

type Config struct {
	Path         string
	IgnoredDirs  []string
	MaxLoopDepth int
	Full         bool
}

func New(config *Config) *Finder {
	if config == nil {
		config = &Config{}
	}

	if config.Path == "" {
		config.Path = "."
	}

	return &Finder{
		path:         config.Path,
		ignoredDirs:  config.IgnoredDirs,
		maxLoopDepth: config.MaxLoopDepth,
		full:         config.Full,
	}
}

func (f *Finder) FindProjects() ([]*project.Project, error) {
	var projects []*project.Project
	errorChan := make(chan error, 1)
	pathChan := f.walkFolder(f.path, errorChan)
	projectChan := scanProject(f.full, pathChan)

	for {
		select {
		case err := <-errorChan:
			return nil, fmt.Errorf("FindProjects: %w", err)
		case prj, ok := <-projectChan:
			if !ok {
				return projects, nil
			}
			projects = append(projects, prj)
		}
	}
}

// func walkFolder(filePath string) ([]*project.Project, error) {
func (f *Finder) walkFolder(filePath string, errorChan chan<- error) <-chan string {
	findingChan := make(chan string)
	visited := make(map[string]bool)

	walkFn := func(currentPath string, info fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", currentPath, err)
		}

		absPath := helpers.AbsolutePath(currentPath)
		if visited[absPath] {
			return filepath.SkipDir
		}
		visited[absPath] = true

		if info.IsDir() {
			if slices.Contains(f.ignoredDirs, info.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		base := filepath.Base(currentPath)

		switch base {
		case "go.mod":
			findingChan <- absPath
			return filepath.SkipDir
		case "go.work":
			modPathList, err := helpers.ParseWorkspaceFile(absPath, absPath)
			if err != nil {
				errorChan <- err
				return filepath.SkipDir
			}

			for _, path := range modPathList {
				findingChan <- path
			}

			return filepath.SkipDir
		default:
			return nil
		}
	}

	go func() {
		if err := filepath.WalkDir(filePath, walkFn); err != nil {
			errorChan <- err
		}
		close(findingChan)
	}()

	return findingChan
}

func scanProject(full bool, pathChan <-chan string) <-chan *project.Project {
	projectChan := make(chan *project.Project)

	go func() {
		for {
			path, ok := <-pathChan
			if !ok {
				close(projectChan)
				break
			}

			project := helpers.ParseModFile(path)

			if project.Error == nil {
				project.Scan(full)
			}

			projectChan <- project
		}
	}()

	return projectChan
}
