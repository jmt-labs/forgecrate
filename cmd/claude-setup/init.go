// cmd/claude-setup/init.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profile string
	var flavors []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialisiert Claude-Setup im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("init: profile=%s flavors=%v\n", profile, flavors)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "backend", "Profil (backend|frontend|fullstack)")
	cmd.Flags().StringSliceVar(&flavors, "flavors", nil, "Flavors (tdd,strict-review,minimal)")
	return cmd
}
