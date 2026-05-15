/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/padiazg/test-finder/internal/project"
	"github.com/padiazg/test-finder/internal/scan"
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
	scanCmd.Flags().Bool("full", false, "Include fully covered functions")
	scanCmd.Flags().Duration("timeout", 5*time.Minute, "Timeout for scanning operations")
}

func scanCmdFn(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat = "table" // default
	}

	full, _ := cmd.Flags().GetBool("full")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigChan:
			fmt.Fprintf(os.Stderr, "\nreceived signal %s, starting shutown\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	finder := scan.New(&scan.Config{
		Path:        path,
		IgnoredDirs: helpers.IgnoredDirs(),
		Full:        full,
	})

	projects, err := finder.FindProjects(ctx)

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		fmt.Fprintf(os.Stderr, "cancelled: timeout")
		os.Exit(2)
	case ctx.Err() == context.Canceled:
		fmt.Fprintf(os.Stderr, "cancelled: signal received")
		os.Exit(130)
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	switch outputFormat {
	case "table":
		outputTable(projects)
	case "json":
		if err := outputJSON(projects); err != nil {
			return err
		}
	default:
		fmt.Printf("Unknown output format: %s. Using table.", outputFormat)
		outputTable(projects)
	}

	return nil
}

func outputTable(projects []*project.Project) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"Project", "Package", "File", "Function", "Coverage"})

	for _, prj := range projects {
		if len(prj.Files) == 0 {
			continue
		}

		for _, file := range prj.Files {
			for i, fn := range file.Coverage {
				var (
					projectPath string
					fileName    string
					packageName string
				)

				if i == 0 && file.Coverage[0].Name == fn.Name {
					projectPath = prj.Path
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

	t.Render()

	for _, prj := range projects {
		if len(prj.Warnings) == 0 {
			continue
		}
		fmt.Printf("%s\n", prj.Module)
		for _, warn := range prj.Warnings {
			fmt.Printf("  %s\n", warn)
		}
	}
}

func outputJSON(projects []*project.Project) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(projects); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)

	}

	return nil
}
