// cmd/forgecrate/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if filepath.Base(os.Args[0]) == "claude-setup" {
		fmt.Fprintln(os.Stderr, "Warning: 'claude-setup' is deprecated, use 'forgecrate' instead.")
	}
	root := &cobra.Command{
		Use:     "forgecrate",
		Short:   "Reproducible Claude Code configuration for Git repositories.",
		Version: version,
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newDescribeCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newHookCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
