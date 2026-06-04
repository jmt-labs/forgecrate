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

	var warnings []string
	for _, p := range ext.Plugins {
		var cmd *exec.Cmd
		if p.Method == "marketplace" {
			cmd = exec.Command(claude, "plugin", "marketplace", "add", p.Source)
		} else {
			cmd = exec.Command(claude, "plugin", "install", "--scope", "project", p.Source)
		}
		cmd.Dir = i.Dir
		if cmdOut, err := cmd.CombinedOutput(); err != nil {
			msg := string(cmdOut)
			var reason string
			if strings.Contains(msg, "not found in any configured marketplace") {
				reason = "not found in marketplace"
			} else {
				reason = err.Error()
			}
			_, _ = fmt.Fprintf(out, "⚠️  plugin:%s  (%s) — skipped\n", p.Name, reason)
			warnings = append(warnings, fmt.Sprintf("plugin:%s (%s)", p.Name, reason))
			continue
		}
		_, _ = fmt.Fprintf(out, "✅ plugin:%s\n", p.Name)
	}

	if len(warnings) > 0 {
		_, _ = fmt.Fprintf(out, "Hinweis: %d Plugin(s) konnten nicht installiert werden: %s\n",
			len(warnings), strings.Join(warnings, ", "))
	}
	return nil
}

