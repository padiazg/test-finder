/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/padiazg/test-finder/internal/scan"
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
}

func scanCmdFn(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	projects, err := scan.FindProjects(&scan.FindProjectsOpts{Path: path, MaxDepth: 0})
	if err != nil {
		return fmt.Errorf("scan finding projects: %w", err)
	}

	b, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("scan marshalling results: %w", err)
	}

	fmt.Printf("projects: %v", string(b))

	return nil
}
