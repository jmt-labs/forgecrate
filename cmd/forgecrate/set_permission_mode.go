package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	"github.com/spf13/cobra"
)

func newSetPermissionModeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-permission-mode <mode>",
		Short: "Setzt den Agent-Berechtigungsmodus (bypass|plan|ask|auto)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := args[0]
			if err := config.ValidatePermissionMode(mode); err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := filepath.Join(cwd, ".forgecrate.yaml")
			cfg, err := config.Read(cfgPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf(".forgecrate.yaml nicht gefunden — erst 'forgecrate init' ausführen")
			} else if err != nil {
				return err
			}

			cfg.PermissionMode = mode

			if err := deploy.PatchPermissionMode(cwd, mode, &cfg); err != nil {
				return err
			}

			if err := config.Write(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("✓ permission_mode: %s\n", mode)
			fmt.Println("✓ .claude/settings.json aktualisiert")
			return nil
		},
	}
}
