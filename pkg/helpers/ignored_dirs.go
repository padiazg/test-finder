package helpers

func IgnoredDirs() []string {
	return []string{
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
}
