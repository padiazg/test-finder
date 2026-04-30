package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"

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

type findingPair struct {
	err  error
	path string
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
	var (
		projects []*project.Project
		wg       sync.WaitGroup
	)
	findingChan := f.walkFolder(f.path)
	projectChan := make(chan *project.Project, 10)

	wg.Go(func() {
		for finding := range findingChan {
			if finding.err != nil {
				projectChan <- &project.Project{Error: finding.err}
				continue
			}

			wg.Go(func() { f.parseMod(finding.path, projectChan) })
		}
	})

	go func() {
		wg.Wait()
		close(projectChan)
	}()

	for prj := range projectChan {
		// fmt.Printf("> prj: %s err: %v\n", prj.Module, prj.Error)
		projects = append(projects, prj)
	}

	return projects, nil
}

// func walkFolder(filePath string) ([]*project.Project, error) {
func (f *Finder) walkFolder(filePath string) <-chan findingPair {
	findingChan := make(chan findingPair)

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
			findingChan <- findingPair{path: absPath}
			return filepath.SkipDir
		case "go.work":
			modPathList, err := helpers.ParseWorkspaceFile(absPath)
			if err != nil {
				findingChan <- findingPair{path: absPath, err: err}
				return filepath.SkipDir
			}

			for _, path := range modPathList {
				findingChan <- findingPair{path: path}
			}

			return filepath.SkipDir
		default:
			return nil
		}
	}

	go func() {
		if err := filepath.WalkDir(filePath, walkFn); err != nil {
			findingChan <- findingPair{path: filePath, err: err}
		}
		close(findingChan)
	}()

	return findingChan
}

func (f *Finder) parseMod(path string, projectChan chan<- *project.Project) {
	prj := helpers.ParseModFile(path)
	if prj.Error == nil {
		prj.Scan(f.full)
	}

	projectChan <- prj
}
