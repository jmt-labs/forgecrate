package extensions

import (
	"log"
	"os"
	"os/exec"
)

type Installer struct {
	Claude string
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
		cmd := exec.Command(claude, "plugin", "install", p.Source)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("warn: plugin install %s: %v: %s", p.Name, err, out)
		}
	}

	for _, m := range ext.MCP {
		scope := m.Scope
		if scope == "" {
			scope = "local"
		}
		args := []string{"mcp", "add", m.Name, "--scope", scope, m.Command}
		args = append(args, m.Args...)
		cmd := exec.Command(claude, args...)
		cmd.Env = append(os.Environ(), envPairs(m.Env)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("warn: mcp add %s: %v: %s", m.Name, err, out)
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
