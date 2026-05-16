package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	gh "github.com/jmt-labs/claude-setup/internal/github"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Listet verfügbare Profile und Flavors auf",
		RunE: func(cmd *cobra.Command, args []string) error {
			srcDir, err := os.MkdirTemp("", "claude-setup-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(srcDir)

			fmt.Println("Fetching jmt-labs/claude-setup@main ...")
			client := gh.Default()
			if err := client.Download("jmt-labs", "claude-setup", "main", srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			profiles, err := listDirs(filepath.Join(srcDir, "profiles"))
			if err != nil {
				return err
			}
			flavors, err := listDirs(filepath.Join(srcDir, "flavors"))
			if err != nil {
				return err
			}

			fmt.Println("\nProfile:")
			for _, p := range profiles {
				if p == "backend" {
					fmt.Printf("  %-20s (Standard)\n", p)
				} else {
					fmt.Printf("  %s\n", p)
				}
			}
			fmt.Println("\nFlavors:")
			for _, f := range flavors {
				fmt.Printf("  %s\n", f)
			}
			return nil
		},
	}
}

func listDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
