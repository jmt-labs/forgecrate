// cmd/claude-setup/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "claude-setup",
		Short:   "Reproduzierbares Claude-Setup für Repos",
		Version: version,
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newDescribeCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
