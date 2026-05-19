package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profile string
	var flavors []string

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"run"},
		Short:   "Initialisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := cwd + "/.forgecrate.yaml"
			cfg, err := config.Read(cfgPath)
			var legacyPath string
			if errors.Is(err, os.ErrNotExist) {
				legacy := cwd + "/.claude-setup.yaml"
				if legacyCfg, lerr := config.Read(legacy); lerr == nil {
					cfg = legacyCfg
					legacyPath = legacy
					fmt.Fprintln(os.Stderr, "Notice: migrating .claude-setup.yaml → .forgecrate.yaml")
				} else {
					cfg = config.Config{
						Version: "1.0",
						Source:  "github.com/jmt-labs/forgecrate",
						Ref:     "main",
						Profile: profile,
						Flavors: flavors,
					}
				}
			} else if err != nil {
				return err
			}

			if cmd.Flags().Changed("profile") {
				cfg.Profile = profile
			}
			if len(flavors) > 0 {
				cfg.Flavors = flavors
			}

			owner, repo := "jmt-labs", "forgecrate"
			fmt.Printf("Fetching %s/%s@%s ...\n", owner, repo, cfg.Ref)

			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(srcDir) }()

			client := gh.Default()
			if err := client.Download(owner, repo, cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("Deploying profile=%s flavors=%v ...\n", cfg.Profile, cfg.Flavors)
			if err := deploy.Run(srcDir, cwd, cfg); err != nil {
				return err
			}

			if legacyPath != "" {
				_ = os.Remove(legacyPath)
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "backend", "Profil (backend|frontend|fullstack)")
	cmd.Flags().StringSliceVar(&flavors, "flavors", nil, "Flavors (tdd,strict-review,minimal)")
	return cmd
}
