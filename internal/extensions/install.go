package extensions

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
)

// Installer invokes the claude CLI to install plugins defined in an Extensions set.
type Installer struct {
	Claude string
	Dir    string    // git repo root of the target project
	Out    io.Writer // progress output; nil = silent
	Scope  string    // plugin scope: "project" (default) | "user"; empty ⇒ "project"
	DryRun bool      // when true, print commands instead of executing them
}

// NewInstaller returns an Installer that uses the "claude" binary on PATH.
func NewInstaller() Installer {
	return Installer{Claude: "claude"}
}

func (i Installer) claudeBin() string {
	if i.Claude == "" {
		return "claude"
	}
	return i.Claude
}

func (i Installer) out() io.Writer {
	if i.Out == nil {
		return io.Discard
	}
	return i.Out
}

func (i Installer) Install(ext Extensions) error {
	claude := i.claudeBin()
	out := i.out()

	scope := i.Scope
	if scope == "" {
		scope = "project"
	}

	for _, p := range ext.Plugins {
		var cmd *exec.Cmd
		if p.Method == "marketplace" {
			cmd = exec.Command(claude, "plugin", "marketplace", "add", p.Source)
		} else {
			cmd = exec.Command(claude, "plugin", "install", "--scope", scope, p.Source)
		}
		cmd.Dir = i.Dir

		if i.DryRun {
			_, _ = fmt.Fprintf(out, "🟡 dry-run plugin:%s  (%s)\n", p.Name, strings.Join(cmd.Args, " "))
			continue
		}

		if cmdOut, err := cmd.CombinedOutput(); err != nil {
			msg := string(cmdOut)
			if strings.Contains(msg, "already installed") {
				_, _ = fmt.Fprintf(out, "🔵 plugin:%s  (already installed)\n", p.Name)
				continue
			}
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

// InstallMCPUser registers MCP servers at user (host) scope via `claude mcp add`.
// It is idempotent: servers already registered (detected via `claude mcp get`)
// are skipped instead of erroring.
func (i Installer) InstallMCPUser(ext Extensions) error {
	claude := i.claudeBin()
	out := i.out()

	for _, m := range ext.MCP {
		if !i.DryRun && i.mcpExists(m.Name) {
			_, _ = fmt.Fprintf(out, "🔵 mcp:%s  (already registered)\n", m.Name)
			continue
		}

		cmd := exec.Command(claude, mcpAddArgs(m)...)

		if i.DryRun {
			_, _ = fmt.Fprintf(out, "🟡 dry-run mcp:%s  (%s)\n", m.Name, strings.Join(cmd.Args, " "))
			continue
		}

		if cmdOut, err := cmd.CombinedOutput(); err != nil {
			_, _ = fmt.Fprintf(out, "❌ mcp:%s  (%v)\n", m.Name, err)
			return fmt.Errorf("mcp add %s: %w: %s", m.Name, err, strings.TrimSpace(string(cmdOut)))
		}
		_, _ = fmt.Fprintf(out, "✅ mcp:%s\n", m.Name)
	}

	return nil
}

// mcpExists reports whether an MCP server is already registered.
func (i Installer) mcpExists(name string) bool {
	cmd := exec.Command(i.claudeBin(), "mcp", "get", name)
	return cmd.Run() == nil
}

// mcpAddArgs builds the `claude mcp add --scope user ...` argument list for one
// MCP server. http servers use --transport http <name> <url>; stdio servers use
// [--env K=V ...] <name> -- <command> <args...>.
func mcpAddArgs(m MCP) []string {
	args := []string{"mcp", "add", "--scope", "user"}
	if m.Transport == "http" {
		args = append(args, "--transport", "http", m.Name, m.URL)
		return args
	}

	keys := make([]string, 0, len(m.Env))
	for k := range m.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		args = append(args, "--env", k+"="+m.Env[k])
	}

	args = append(args, m.Name, "--", m.Command)
	args = append(args, m.Args...)
	return args
}
