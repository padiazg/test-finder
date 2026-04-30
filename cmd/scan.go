/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/padiazg/test-finder/internal/project"
	"github.com/padiazg/test-finder/internal/scan/v2"
	"github.com/padiazg/test-finder/pkg/helpers"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines and likely contains examples`,
	RunE:  scanCmdFn,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringP("output", "o", "table", "Output format (json or table)")
	scanCmd.Flags().Bool("full", false, "Include fully covereded functions")
}

func scanCmdFn(cmd *cobra.Command, args []string) error {
	// Get flags
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat = "table" // default
	}

	full, _ := cmd.Flags().GetBool("full")

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	finder := scan.New(&scan.Config{
		Path:        path,
		IgnoredDirs: helpers.IgnoredDirs(),
		Full:        full,
	})

	projects, err := finder.FindProjects()
	if err != nil {
		return fmt.Errorf("scan finding projects: %w", err)
	}

	switch outputFormat {
	case "table":
		outputTable(projects)
	case "json":
		outputJSON(projects)
	default:
		fmt.Printf("Unknown output format: %s. Using table.\n", outputFormat)
		outputTable(projects)

	}

	return nil
}

func outputTable(projects []*project.Project) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Project", "Package", "File", "Function", "Coverage"})

	for _, project := range projects {
		if len(project.Files) == 0 {
			continue
		}

		for _, file := range project.Files {
			for i, fn := range file.Coverage {
				var (
					projectPath string
					fileName    string
					packageName string
				)

				if i == 0 && file.Coverage[0].Name == fn.Name {
					projectPath = project.Path
					packageName = file.Package
				}

				if i == 0 {
					fileName = file.FileName
				}

				t.AppendRow(table.Row{projectPath, packageName, fileName, fn.Name,
					fmt.Sprintf("%.1f%%", fn.Coverage)})

			}
			// Add a separator row between files (optional)
			t.AppendSeparator()
		}
		// Add a separator row between projects (optional)
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func outputJSON(projects []*project.Project) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(projects); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
