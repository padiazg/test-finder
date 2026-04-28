package scan

type Project struct {
	// Name    string
	Package string
	Path    string
	Files   []FileNode
}

type FileNode struct {
	Path     string
	Coverage []FunctionCoverage
}

type FunctionCoverage struct {
	Name     string
	Line     int
	Coverage float64
}

func (p *Project) Scan() error {

	return nil
}
