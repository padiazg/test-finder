package scan

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/padiazg/test-finder/internal/project"
	"github.com/padiazg/test-finder/pkg/helpers"
	scanerrors "github.com/padiazg/test-finder/pkg/scan_errors"
	"golang.org/x/mod/modfile"
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

func (f *Finder) FindProjects(ctx context.Context) ([]*project.Project, error) {
	var (
		projects []*project.Project
		wg       sync.WaitGroup
	)
	projectChan := make(chan *project.Project, 1)
	findingChan := f.walkFolder(ctx, f.path)

	wg.Go(func() {
		for {
			select {
			case finding, ok := <-findingChan:
				if !ok {
					return
				}
				if finding.err != nil {
					projectChan <- &project.Project{Path: finding.path, Error: finding.err}
					return
				}

				wg.Go(func() {
					prj := f.parseModFile(ctx, finding.path).Scan(ctx, f.full)
					select {
					case projectChan <- prj:
					case <-ctx.Done():
					}
				})
			case <-ctx.Done():
				fmt.Printf("timeout 1\n")
				return
			}
		}
	})

	go func() {
		wg.Wait()
		close(projectChan)
	}()

	for {
		select {
		case prj, ok := <-projectChan:
			if !ok {
				return projects, nil
			}

			if prj.Error != nil {
				if scanerrors.IsError(prj.Error) {
					fmt.Printf("   err: %v\n", prj.Error)
					return nil, prj.Error
				}
				fmt.Printf("   warn: %v\n", prj.Error)
			}
			projects = append(projects, prj)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (f *Finder) walkFolder(ctx context.Context, filePath string) <-chan findingPair {
	findingChan := make(chan findingPair)
	visited := make(map[string]bool)

	walkFn := func(currentPath string, info fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			fmt.Printf("timeout 2\n")
			return &scanerrors.ScanTimeout{Err: fmt.Errorf("walkFn: %w", ctx.Err())}
		default:
		}

		if err != nil {
			return &scanerrors.ScanError{Err: fmt.Errorf("%s: %w", currentPath, err)}
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
			select {
			case findingChan <- findingPair{path: absPath}:
			case <-ctx.Done():
				return &scanerrors.ScanTimeout{Err: fmt.Errorf("walkFn: %w", ctx.Err())}
			}

			return filepath.SkipDir

		// go.work is a Go workspace file that groups multiple modules into a single workspace.
		// It declares modules root via "use" directives. We parse it here so each listed module
		// gets treated as a separate project root to scan, without having to descend into the
		// directory and find go.mod inside
		case "go.work":
			modPathList, err := helpers.ParseWorkspaceFile(absPath)
			if err != nil {
				select {
				case findingChan <- findingPair{path: absPath, err: err}:
				case <-ctx.Done():
					return &scanerrors.ScanTimeout{Err: fmt.Errorf("walkFn: %w", ctx.Err())}
				}

				return filepath.SkipDir
			}

			for _, path := range modPathList {
				select {
				case findingChan <- findingPair{path: path}:
				case <-ctx.Done():
					return &scanerrors.ScanTimeout{Err: fmt.Errorf("walkFn: %w", ctx.Err())}
				}
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

func (f *Finder) parseModFile(ctx context.Context, filePath string) *project.Project {
	absFilePath := helpers.AbsolutePath(filePath)
	prj := &project.Project{Path: filepath.Dir(absFilePath)}

	select {
	case <-ctx.Done():
		prj.Error = &scanerrors.ScanTimeout{Err: fmt.Errorf("walkFn: %w", ctx.Err())}
		return prj
	default:
	}

	fileData, err := os.ReadFile(absFilePath)
	if err != nil {
		prj.Error = fmt.Errorf("reading %s: %w", absFilePath, err)
		return prj
	}

	file, err := modfile.Parse(absFilePath, fileData, nil)
	if err != nil {
		prj.Error = fmt.Errorf("parsing %s: %w", absFilePath, err)
		return prj
	}

	if file.Module.Mod.Path == "" {
		prj.Error = fmt.Errorf("no module readed from %s", absFilePath)
		return prj
	}

	prj.Module = file.Module.Mod.Path

	return prj
}
