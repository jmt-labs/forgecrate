package extensions

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Installer invokes the claude CLI to install plugins defined in an Extensions set.
type Installer struct {
	Claude string
	Dir    string    // git repo root of the target project
	Out    io.Writer // progress output; nil = silent
}

// NewInstaller returns an Installer that uses the "claude" binary on PATH.
func NewInstaller() Installer {
	return Installer{Claude: "claude"}
}

func (i Installer) Install(ext Extensions) error {
	claude := i.Claude
	if claude == "" {
		claude = "claude"
	}
	out := i.Out
	if out == nil {
		out = io.Discard
	}

	for _, p := range ext.Plugins {
		if isPluginInstalled(claude, i.Dir, p.Name) {
			_, _ = fmt.Fprintf(out, "🔵 plugin:%s  (bereits installiert)\n", p.Name)
			continue
		}

		var cmd *exec.Cmd
		if p.Method == "marketplace" {
			cmd = exec.Command(claude, "plugin", "marketplace", "add", p.Source)
		} else {
			cmd = exec.Command(claude, "plugin", "install", "--scope", "project", p.Source)
		}
		cmd.Dir = i.Dir
		if cmdOut, err := cmd.CombinedOutput(); err != nil {
			msg := string(cmdOut)
			if strings.Contains(msg, "not found in any configured marketplace") {
				_, _ = fmt.Fprintf(out, "❌ plugin:%s  (not found in marketplace)\n", p.Name)
				return fmt.Errorf("plugin %s not found in marketplace: %w", p.Name, err)
			}
			_, _ = fmt.Fprintf(out, "❌ plugin:%s  (%v)\n", p.Name, err)
			return fmt.Errorf("plugin install %s: %w", p.Name, err)
		}
		_, _ = fmt.Fprintf(out, "✅ plugin:%s\n", p.Name)
	}

	return nil
}

// isPluginInstalled prüft via "claude plugin list" ob ein Plugin bereits installiert ist.
// Bei Fehler (claude nicht gefunden etc.) wird false zurückgegeben (fail-open).
func isPluginInstalled(claude, dir, name string) bool {
	cmd := exec.Command(claude, "plugin", "list")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), name)
}

