// cmd/claude-setup/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "claude-setup",
		Short: "Reproduzierbares Claude-Setup für Repos",
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newListCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
