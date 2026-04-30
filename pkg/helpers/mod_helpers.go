package helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/padiazg/test-finder/internal/project"
	"golang.org/x/mod/modfile"
)

func ParseModFile(filePath string) *project.Project {
	absFilePath := AbsolutePath(filePath)
	project := &project.Project{Path: filepath.Dir(absFilePath)}

	fileData, err := os.ReadFile(absFilePath)
	if err != nil {
		project.Error = fmt.Errorf("reading %s: %w", absFilePath, err)
		return project
	}

	file, err := modfile.Parse(absFilePath, fileData, nil)
	if err != nil {
		project.Error = fmt.Errorf("parsing %s: %w", absFilePath, err)
		return project
	}

	if file.Module.Mod.Path == "" {
		project.Error = fmt.Errorf("no module readed from %s", absFilePath)
		return project
	}

	project.Module = file.Module.Mod.Path

	return project
}

func ParseWorkspaceFile(workspacePath string, basePath string) ([]string, error) {
	absWorkspacePath := AbsolutePath(workspacePath)
	absBasePath := filepath.Dir(absWorkspacePath)

	fileData, err := os.ReadFile(absWorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("reading workspace %s: %w", absWorkspacePath, err)
	}

	file, err := modfile.ParseWork(absWorkspacePath, fileData, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace %s, %w", absWorkspacePath, err)
	}

	var modPathList []string

	if file.Use == nil {
		return modPathList, nil
	}

	for _, directive := range file.Use {
		dirPath := filepath.Clean(directive.Path)
		modPath := filepath.Join(absBasePath, dirPath, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			modPathList = append(modPathList, modPath)
		}
	}

	return modPathList, nil
}
