package extensions

import (
	"fmt"
	"io"
	"os"
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
				fmt.Fprintf(out, "❌ plugin:%s  (not found in marketplace)\n", p.Name)
				return fmt.Errorf("plugin %s: not found in marketplace", p.Name)
			}
			fmt.Fprintf(out, "❌ plugin:%s  (%v)\n", p.Name, err)
			return fmt.Errorf("plugin install %s: %w", p.Name, err)
		}
		fmt.Fprintf(out, "✅ plugin:%s\n", p.Name)
	}

	for _, m := range ext.MCP {
		var args []string
		if m.Transport == "http" {
			args = []string{"mcp", "add", "--transport", "http", m.Name, m.URL, "--scope", "project"}
		} else {
			args = []string{"mcp", "add", m.Name, "--scope", "project", m.Command}
			if len(m.Args) > 0 {
				args = append(args, "--")
			}
			args = append(args, m.Args...)
		}

		cmd := exec.Command(claude, args...)
		cmd.Dir = i.Dir
		cmd.Env = append(os.Environ(), envPairs(m.Env)...)
		if cmdOut, err := cmd.CombinedOutput(); err != nil {
			msg := string(cmdOut)
			if strings.Contains(msg, "already exists") {
				fmt.Fprintf(out, "🔵 mcp:%s  (already configured)\n", m.Name)
				continue
			}
			fmt.Fprintf(out, "❌ mcp:%s  (%v)\n", m.Name, err)
			return fmt.Errorf("mcp add %s: %w", m.Name, err)
		}
		fmt.Fprintf(out, "✅ mcp:%s\n", m.Name)
	}
	return nil
}

func envPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for k, v := range env {
		pairs = append(pairs, k+"="+v)
	}
	return pairs
}
