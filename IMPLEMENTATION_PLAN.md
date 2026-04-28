# FindProjects Implementation Plan

Function: `FindProjects(path string, opts *FindProjectsOpts) ([]Project, error)`

## Architecture

### Directory Structure
```
internals/scan/
├── scan.go        # Core implementation
└── scan_test.go   # Unit tests (placeholder)
```

### Project Struct
```go
type Project struct {
    Name    string  // Module path or directory basename
    Path    string  // Absolute path
    Package string  // Module path from go.mod or go.work
}
```

## Implementation Steps

### [x] Step 1: Define FindProjectsOpts
- Create options struct with MaxDepth field (0 = unlimited)
- Provide default constructor
- Add safety constants (max symlink loops, etc.)

### [x] Step 2: Implement getModulePath(filePath string) -> string
- Parse go.mod file using golang.org/x/mod/modfile
- Extract Module.Mod.Path
- Handle read errors, return "" on failure
- Ensure filepath is absolute before parsing

### [x] Step 3: Implement parseWorkspaceEntries(workspacePath string, basePath string) -> []Project
- Read go.work file
- Parse use directives (lines starting with "./")
- For each ./dir entry, check if dir/go.mod exists
- Parse and collect module paths
- Return slice of Project structs

### [x] Step 4: Implement walkDirFn(path string, info os.FileInfo, err error) filepath.WalkDirFunc
- Skip ignored directories (.git, vendor, node_modules, etc.)
- Track directory depth
- Check MaxDepth limit if not 0
- Skip directories already visited (symlink loop detection)
- If go.mod or go.work found:
  - Create Project entry
  - Skip parsing this directory further (no nested markers)

### [x] Step 5: Implement getMarkerFileCount(path string) -> (int, int)
- Walk path, count go.mod and go.work files
- Ignore ignored directories
- Return count of each type

### [x] Step 6: Core FindProjects logic
- Path resolution
- Marker count check
- Handle single marker
- Handle workspace
- Handle walk path

### [x] Step 7: Error handling
- Return ([]Project{}, nil) on any error
- Skip directories with permission issues
- Skip malformed files
- Symlink loop detection

### [ ] Step 8: Integration
- Update cmd/scan.go to use FindProjects
- Add command line flags for MaxDepth
- Add logging/debug output

## Tests
### [ ] Step 9: Unit tests
- Test single go.mod
- Test go.work workspace
- Test multiple projects
- Test ignore list
- Test max depth
- Test symlink loop
- Test error cases

## Documentation
### [ ] Step 10: Update module documentation
- Add doc comments for Project struct
- Add doc comments for FindProjects
- Add usage examples in cmd/scan.go Long field

## Status
[ ] Plan written and approved
[ ] Step 1 - 4 implemented
[ ] Step 5 - 6 implemented
[ ] Step 7 - 8 implemented
[ ] Testing complete
[ ] Documentation updated
[ ] Code review complete
[ ] Ready for merge