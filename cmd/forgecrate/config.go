package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

type promptFn func(profiles, flavors []string, cur config.Config) (string, []string, error)

func configInteractive(cwd, srcDir string, cfg config.Config, prompt promptFn) (config.Config, error) {
	profiles, err := listDirs(filepath.Join(srcDir, "profiles"))
	if err != nil {
		return config.Config{}, fmt.Errorf("profile-Liste lesen: %w", err)
	}
	if len(profiles) == 0 {
		return config.Config{}, fmt.Errorf("keine Profile im Source-Repo gefunden")
	}

	flavors, err := listDirs(filepath.Join(srcDir, "flavors"))
	if err != nil {
		return config.Config{}, fmt.Errorf("flavor-Liste lesen: %w", err)
	}
	if len(flavors) == 0 {
		return config.Config{}, fmt.Errorf("keine Flavors im Source-Repo gefunden")
	}

	newProfile, newFlavors, err := prompt(profiles, flavors, cfg)
	if err != nil {
		return config.Config{}, err
	}

	cfg.Profile = newProfile
	cfg.Flavors = newFlavors
	if err := config.Write(filepath.Join(cwd, ".forgecrate.yaml"), cfg); err != nil {
		return config.Config{}, fmt.Errorf("config schreiben: %w", err)
	}
	return cfg, nil
}

func huhPrompt(profiles, flavors []string, cur config.Config) (string, []string, error) {
	newProfile := cur.Profile
	newFlavors := make([]string, len(cur.Flavors))
	copy(newFlavors, cur.Flavors)

	profileOpts := make([]huh.Option[string], len(profiles))
	for i, p := range profiles {
		profileOpts[i] = huh.NewOption(p, p)
	}

	flavorOpts := make([]huh.Option[string], len(flavors))
	for i, f := range flavors {
		flavorOpts[i] = huh.NewOption(f, f)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Profil").
				Options(profileOpts...).
				Value(&newProfile),
			huh.NewMultiSelect[string]().
				Title("Flavors  (Leertaste = toggle)").
				Options(flavorOpts...).
				Value(&newFlavors),
		),
	)

	if err := form.Run(); err != nil {
		return "", nil, err
	}
	return newProfile, newFlavors, nil
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Profil und Flavors interaktiv konfigurieren",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := filepath.Join(cwd, ".forgecrate.yaml")
			cfg, err := config.Read(cfgPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("kein forgecrate-Repo. Zuerst `forgecrate init` ausführen")
			}
			if err != nil {
				return err
			}

			fmt.Printf("Fetching jmt-labs/forgecrate@%s ...\n", cfg.Ref)
			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(srcDir) }()

			client := gh.Default()
			if err := client.Download("jmt-labs", "forgecrate", cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			updatedCfg, err := configInteractive(cwd, srcDir, cfg, huhPrompt)
			if err != nil {
				return err
			}

			fmt.Printf("Deploying profile=%s flavors=%v ...\n", updatedCfg.Profile, updatedCfg.Flavors)
			if err := deploy.Run(srcDir, cwd, updatedCfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
}
