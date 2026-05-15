package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	scanerrors "github.com/padiazg/test-finder/pkg/scan_errors"
)

type Project struct {
	Error    error
	Module   string
	Path     string
	Files    []FileNode
	Warnings []string
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

func (p *Project) Scan(ctx context.Context, full bool) *Project {
	coverFilePath := filepath.Join(p.Path, "coverage.out")
	testCmd := exec.CommandContext(ctx, "go", "test", "-coverprofile=coverage.out", "-cover", "./...")
	testCmd.Dir = p.Path

	// check if there is a coverage file
	if _, err := os.Stat(coverFilePath); err == nil {
		if err := os.Remove(coverFilePath); err != nil {
			p.Error = &scanerrors.ScanWarning{Err: fmt.Errorf("unable to remove coverage.out: %w", err)}
			return p
		}
	}

	// the test count fail partially and still produce a coverage file we can use
	if err := testCmd.Run(); err != nil {
		p.Warnings = append(p.Warnings, fmt.Sprintf("go test failed: %v\n", err))
	}

	// cover file not found, can't continue
	if _, err := os.Stat(coverFilePath); err != nil {
		p.Error = &scanerrors.ScanWarning{Err: fmt.Errorf("cover file: %w", err)}
		return p
	}

	defer func() {
		if err := os.Remove(coverFilePath); err != nil {
			p.Warnings = append(p.Warnings, fmt.Sprintf("removing %s", coverFilePath))
		}
	}()

	coverCmd := exec.CommandContext(ctx, "go", "tool", "cover", "-func=coverage.out")
	coverCmd.Dir = p.Path
	out, err := coverCmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			fmt.Printf("cover: %v\n", ctx.Err())
			p.Error = &scanerrors.ScanTimeout{Err: ctx.Err()}
		} else {
			fmt.Printf("cover: %v\n", err)
			p.Error = &scanerrors.ScanWarning{Err: err}
		}
		return p
	}

	re := regexp.MustCompile(`^([^:]+):(\d+):\s+(\S+)\s+([\d.]+)%$`)
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
		if _, err := fmt.Sscanf(matches[2], "%d", &lineNum); err != nil {
			p.Warnings = append(p.Warnings, fmt.Sprintf("%s: parsing line number\n", filePath))

		}

		funcName := matches[3]
		var cov float64
		if _, err := fmt.Sscanf(matches[4], "%f", &cov); err != nil {
			p.Warnings = append(p.Warnings, fmt.Sprintf("%s: parsing coverage\n", filePath))
		}

		if !full && cov == 100.00 {
			continue
		}

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
			if !full && file.Average == 100.0 {
				continue
			}
			p.Files = append(p.Files, *file)
		}

		slices.SortFunc(p.Files, func(a, b FileNode) int {
			if cmp := strings.Compare(a.Package, b.Package); cmp != 0 {
				return cmp
			}

			return strings.Compare(a.FileName, b.FileName)
		})
	}

	return p
}

func (c FunctionCoverageList) Average() float64 {
	q := len(c)

	switch q {
	case 0:
		return 0.0
	case 1:
		return c[0].Coverage
	default:
		var sumCoverage float64
		for _, coverage := range c {
			sumCoverage += coverage.Coverage
		}
		return sumCoverage / float64(q)
	}
}
