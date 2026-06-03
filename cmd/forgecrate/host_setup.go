package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/jmt-labs/forgecrate/internal/extensions"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/jmt-labs/forgecrate/internal/prereqs"
	"github.com/spf13/cobra"
)

type hostSetupOpts struct {
	Scope       string // "host" | "project"
	Yes         bool
	DryRun      bool
	SkipPrereqs bool
	ClaudeBin   string // path to claude binary; empty ⇒ "claude"
	TargetDir   string // for project scope: repo root; for host scope: cwd
}

// scopePrompt asks the user to choose a scope; overridable in tests.
type scopePrompt func() (string, error)

func huhScopePrompt() (string, error) {
	scope := "host"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Scope").
				Options(
					huh.NewOption("Host / user-global (alle Projekte)", "host"),
					huh.NewOption("Projekt (nur dieses Repo)", "project"),
				).
				Value(&scope),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return scope, nil
}

// runHostSetup performs the host-setup using an already-downloaded source dir.
// It is the testable core of the command (no network).
func runHostSetup(srcDir string, opts hostSetupOpts, out io.Writer) error {
	if opts.Scope != "host" && opts.Scope != "project" {
		return fmt.Errorf("ungültiger scope %q (erwartet: host|project)", opts.Scope)
	}

	union, err := extensions.CollectAll(srcDir)
	if err != nil {
		return fmt.Errorf("extensions sammeln: %w", err)
	}
	fmt.Fprintf(out, "Gefunden: %d Plugins, %d MCP-Server (Union über base + alle Profile + alle Flavors)\n",
		len(union.Plugins), len(union.MCP))

	// Prerequisites
	if !opts.SkipPrereqs {
		checker := prereqs.Checker{Out: out, DryRun: opts.DryRun, Assume: opts.Yes}
		if err := checker.EnsureClaude(); err != nil {
			return err
		}
		present, missing := checker.Detect(prereqs.DefaultTools())
		for _, t := range present {
			fmt.Fprintf(out, "✅ %s\n", t.Name)
		}
		for _, t := range missing {
			if t.Name == "codegraph" {
				continue // handled by EnsureCodegraph below
			}
			fmt.Fprintf(out, "⚠️  %s fehlt — %s\n", t.Name, t.Hint)
		}
		if err := checker.EnsureCodegraph(); err != nil {
			return err
		}
	}

	pluginScope := "project"
	if opts.Scope == "host" {
		pluginScope = "user"
	}

	installer := extensions.Installer{
		Claude: opts.ClaudeBin,
		Dir:    opts.TargetDir,
		Out:    out,
		Scope:  pluginScope,
		DryRun: opts.DryRun,
	}

	if err := installer.Install(union); err != nil {
		return err
	}

	if opts.Scope == "host" {
		if err := installer.InstallMCPUser(union); err != nil {
			return err
		}
	} else {
		if opts.DryRun {
			fmt.Fprintf(out, "🟡 dry-run mcp: würde .mcp.json in %s schreiben (%d Server)\n", opts.TargetDir, len(union.MCP))
		} else if err := extensions.WriteMCPJson(opts.TargetDir, union); err != nil {
			return err
		} else if len(union.MCP) > 0 {
			fmt.Fprintf(out, "✅ .mcp.json (%d Server)\n", len(union.MCP))
		}
	}

	fmt.Fprintln(out, "Fertig.")
	return nil
}

func newHostSetupCmd() *cobra.Command {
	var scope string
	var yes, dryRun, skipPrereqs bool
	var ref string

	cmd := &cobra.Command{
		Use:   "host-setup",
		Short: "Richtet diese Maschine für forgecrate + Claude ein (Plugins, MCP-Server, Prerequisites)",
		Long: "Installiert die Union aller Plugins und MCP-Server (base + alle Profile + alle Flavors) " +
			"host-global oder projektweit und installiert fehlende Prerequisites.\n\n" +
			"Hinweis: --yes führt ggf. den remote codegraph install.sh ohne Rückfrage aus.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if scope == "" {
				if yes {
					return fmt.Errorf("--scope (host|project) ist im nicht-interaktiven Modus (--yes) erforderlich")
				}
				s, err := resolveScope(huhScopePrompt)
				if err != nil {
					return err
				}
				scope = s
			}
			if scope != "host" && scope != "project" {
				return fmt.Errorf("ungültiger scope %q (erwartet: host|project)", scope)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if scope == "project" {
				if _, err := os.Stat(filepath.Join(cwd, ".forgecrate.yaml")); os.IsNotExist(err) {
					return fmt.Errorf("kein forgecrate-Repo (kein .forgecrate.yaml). Zuerst `forgecrate init` ausführen oder --scope host nutzen")
				}
			}

			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(srcDir) }()

			fmt.Fprintf(out, "Fetching jmt-labs/forgecrate@%s ...\n", ref)
			client := gh.Default()
			if err := client.Download("jmt-labs", "forgecrate", ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			opts := hostSetupOpts{
				Scope:       scope,
				Yes:         yes,
				DryRun:      dryRun,
				SkipPrereqs: skipPrereqs,
				ClaudeBin:   os.Getenv("CLAUDE_BIN"),
				TargetDir:   cwd,
			}
			return runHostSetup(srcDir, opts, out)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope: host|project (leer ⇒ interaktive Auswahl)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Nicht-interaktiv: keine Rückfragen (CI). Erfordert --scope")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Nur anzeigen, was passieren würde; nichts ausführen")
	cmd.Flags().BoolVar(&skipPrereqs, "skip-prereqs", false, "Prerequisite-Prüfung/-Installation überspringen")
	cmd.Flags().StringVar(&ref, "ref", "main", "forgecrate-Ref für den Source-Download")
	return cmd
}

func resolveScope(prompt scopePrompt) (string, error) {
	return prompt()
}
