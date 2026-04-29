package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	// Name    string
	Module string
	Path   string
	Files  []FileNode
}

type FileNode struct {
	Coverage FunctionCoverageList
	FileName string
	Package  string
	Path     string
	Average  float64
}

type FunctionCoverage struct {
	Name     string
	Line     int
	Coverage float64
}

type FunctionCoverageList []FunctionCoverage

func (p *Project) Scan() error {
	testCmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-cover", "./...")
	testCmd.Dir = p.Path
	if err := testCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: go test failed for project %s: %v\n", p.Path, err)
		return nil
	}
	defer os.Remove(filepath.Join(p.Path, "coverage.out"))

	coverCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	coverCmd.Dir = p.Path
	out, err := coverCmd.Output()
	if err != nil {
		return fmt.Errorf("cover: %w", err)
	}

	re := regexp.MustCompile(`^([^:]+):(\d+):\s+(\S+)\s+([\d.]+)%$`)
	// var files []FileNode
	fileMap := make(map[string]*FileNode)

	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total:") {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		filePath := matches[1]
		var lineNum int
		fmt.Sscanf(matches[2], "%d", &lineNum)
		funcName := matches[3]
		var cov float64
		fmt.Sscanf(matches[4], "%f", &cov)

		if _, ok := fileMap[filePath]; !ok {
			fileMap[filePath] = &FileNode{
				Path:     filePath,
				Package:  filepath.Dir(filePath),
				FileName: filepath.Base(filePath),
			}
		}
		fileMap[filePath].Coverage = append(fileMap[filePath].Coverage, FunctionCoverage{
			Name:     funcName,
			Line:     lineNum,
			Coverage: cov,
		})
	}

	if len(fileMap) > 0 {
		for _, file := range fileMap {
			file.Average = file.Coverage.Average()
			p.Files = append(p.Files, *file)
		}
	}

	return nil
}

func (c FunctionCoverageList) Average() float64 {
	q := len(c)

	switch {
	case q == 0:
		return 0.0
	case q == 1:
		return c[0].Coverage
	default:
		var sumCoverage float64
		for _, coverage := range c {
			sumCoverage += coverage.Coverage
		}
		return sumCoverage / float64(q)
	}
}

func getPackage(basePath, filePath string) string {
	// Safety check: ensure the full path actually starts with the base
	if !strings.HasPrefix(filePath, basePath) {
		return "" // or return an error
	}

	return strings.TrimPrefix(filePath, basePath)
}
