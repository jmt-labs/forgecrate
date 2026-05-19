package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	hook := &cobra.Command{
		Use:   "hook",
		Short: "Hook-Hilfsprogramme für Claude Code",
	}
	hook.AddCommand(newHookPromptSubmitCmd())
	return hook
}

func newHookPromptSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prompt-submit",
		Short: "Gibt die aktive forgecrate-Konfiguration aus (für prompt-submit Hook)",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := promptSubmitOutput(".")
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
}

func promptSubmitOutput(dir string) (string, error) {
	cfgPath := filepath.Join(dir, ".forgecrate.yaml")
	cfg, err := config.Read(cfgPath)
	if err != nil {
		cfgPath = filepath.Join(dir, ".claude-setup.yaml")
		cfg, err = config.Read(cfgPath)
	}

	var profile, flavors string
	if err != nil {
		profile = "unbekannt"
		flavors = "keine"
	} else {
		profile = cfg.Profile
		if profile == "" {
			profile = "unbekannt"
		}
		if len(cfg.Flavors) > 0 {
			flavors = strings.Join(cfg.Flavors, ", ")
		} else {
			flavors = "keine"
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "## forgecrate — Aktive Konfiguration\n")
	fmt.Fprintf(&sb, "Profil: %s | Flavors: %s\n", profile, flavors)
	fmt.Fprintln(&sb)
	fmt.Fprintf(&sb, "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs\n")
	return sb.String(), nil
}
