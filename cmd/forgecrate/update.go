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

func newUpdateCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Aktualisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := config.Read(cwd + "/.forgecrate.yaml")
			var legacyPath string
			if errors.Is(err, os.ErrNotExist) {
				legacy := cwd + "/.claude-setup.yaml"
				if legacyCfg, lerr := config.Read(legacy); lerr == nil {
					cfg = legacyCfg
					legacyPath = legacy
					fmt.Fprintln(os.Stderr, "Notice: migrating .claude-setup.yaml → .forgecrate.yaml")
				} else {
					return fmt.Errorf(".forgecrate.yaml nicht gefunden — erst 'forgecrate init' ausführen")
				}
			} else if err != nil {
				return err
			}

			if profile != "" {
				cfg.Profile = profile
			}

			owner, repo := "jmt-labs", "forgecrate"
			fmt.Printf("Fetching %s/%s@%s ...\n", owner, repo, cfg.Ref)

			srcDir, err := os.MkdirTemp("", "forgecrate-*")
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

			if legacyPath != "" {
				os.Remove(legacyPath)
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profil wechseln (optional)")
	return cmd
}
