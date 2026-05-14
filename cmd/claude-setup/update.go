// cmd/claude-setup/update.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Aktualisiert Claude-Setup im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("update: profile=%s\n", profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profil wechseln (optional)")
	return cmd
}
