package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jmt-labs/claude-setup/internal/config"
	"github.com/jmt-labs/claude-setup/internal/deploy"
	gh "github.com/jmt-labs/claude-setup/internal/github"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Aktualisiert Claude-Setup im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := config.Read(cwd + "/.claude-setup.yaml")
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf(".claude-setup.yaml nicht gefunden — erst 'claude-setup init' ausführen")
			} else if err != nil {
				return err
			}

			if profile != "" {
				cfg.Profile = profile
			}

			owner, repo := "jmt-labs", "claude-setup"
			fmt.Printf("Fetching %s/%s@%s ...\n", owner, repo, cfg.Ref)

			srcDir, err := os.MkdirTemp("", "claude-setup-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(srcDir)

			client := gh.Default()
			if err := client.Download(owner, repo, cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("Updating profile=%s flavors=%v ...\n", cfg.Profile, cfg.Flavors)
			if err := deploy.Run(srcDir, cwd, cfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profil wechseln (optional)")
	return cmd
}
