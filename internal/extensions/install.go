package extensions

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Installer struct {
	Claude string
	Dir    string    // git repo root of the target project
	Out    io.Writer // progress output; nil = silent
}

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
		cmd := exec.Command(claude, "plugin", "install", "--scope", "project", p.Source)
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

