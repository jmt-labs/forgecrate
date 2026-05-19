package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <profile|flavor> <name>",
		Short: "Zeigt eine detaillierte Beschreibung eines Profils oder Flavors",
		Example: `  claude-setup describe profile backend
  claude-setup describe flavor tdd`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, name := args[0], args[1]

			srcDir, err := os.MkdirTemp("", "claude-setup-*")
			if err != nil {
				return err
			}
			defer func() {
				if err := os.RemoveAll(srcDir); err != nil {
					fmt.Fprintf(os.Stderr, "warning: temp-Verzeichnis konnte nicht gelöscht werden %s: %v\n", srcDir, err)
				}
			}()

			fmt.Println("Fetching jmt-labs/claude-setup@main ...")
			client := gh.Default()
			if err := client.Download("jmt-labs", "claude-setup", "main", srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			out, err := describeEntry(srcDir, kind, name)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
}

func describeEntry(srcDir, kind, name string) (string, error) {
	var dir string
	switch kind {
	case "profile":
		dir = filepath.Join(srcDir, "profiles", name)
	case "flavor":
		dir = filepath.Join(srcDir, "flavors", name)
	default:
		return "", fmt.Errorf("unbekannter Typ %q — erlaubt: profile, flavor", kind)
	}

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	content, err := os.ReadFile(claudeMD)
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("%s %q nicht gefunden", kind, name)
	}
	if err != nil {
		return "", fmt.Errorf("CLAUDE.md für %s %q lesen: %w", kind, name, err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "=== %s: %s ===\n\n", strings.ToUpper(kind), name)
	sb.Write(content)

	if kind == "flavor" {
		skills, err := listDirs(filepath.Join(dir, "skills"))
		if err == nil && len(skills) > 0 {
			fmt.Fprintf(&sb, "\nSkills:\n")
			for _, s := range skills {
				fmt.Fprintf(&sb, "  /%s\n", s)
			}
		}
	}

	return sb.String(), nil
}
