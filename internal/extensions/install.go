package extensions

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type Installer struct {
	Claude string
	Dir    string // git repo root of the target project
}

func NewInstaller() Installer {
	return Installer{Claude: "claude"}
}

func (i Installer) Install(ext Extensions) error {
	claude := i.Claude
	if claude == "" {
		claude = "claude"
	}

	for _, p := range ext.Plugins {
		cmd := exec.Command(claude, "plugin", "install", "--scope", "project", p.Source)
		cmd.Dir = i.Dir
		if out, err := cmd.CombinedOutput(); err != nil {
			msg := string(out)
			if !strings.Contains(msg, "not found in any configured marketplace") {
				log.Printf("warn: plugin install %s: %v: %s", p.Name, err, msg)
			}
		}
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
		if out, err := cmd.CombinedOutput(); err != nil {
			msg := string(out)
			if !strings.Contains(msg, "already exists") {
				log.Printf("warn: mcp add %s: %v: %s", m.Name, err, msg)
			}
		}
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
