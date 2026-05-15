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

func newInitCmd() *cobra.Command {
	var profile string
	var flavors []string

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"run"},
		Short:   "Initialisiert Claude-Setup im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := cwd + "/.claude-setup.yaml"
			cfg, err := config.Read(cfgPath)
			if errors.Is(err, os.ErrNotExist) {
				cfg = config.Config{
					Version: "1.0",
					Source:  "github.com/jmt-labs/claude-setup",
					Ref:     "main",
					Profile: profile,
					Flavors: flavors,
				}
			} else if err != nil {
				return err
			}

			if profile != "" {
				cfg.Profile = profile
			}
			if len(flavors) > 0 {
				cfg.Flavors = flavors
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

			fmt.Printf("Deploying profile=%s flavors=%v ...\n", cfg.Profile, cfg.Flavors)
			if err := deploy.Run(srcDir, cwd, cfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "backend", "Profil (backend|frontend|fullstack)")
	cmd.Flags().StringSliceVar(&flavors, "flavors", nil, "Flavors (tdd,strict-review,minimal)")
	return cmd
}
