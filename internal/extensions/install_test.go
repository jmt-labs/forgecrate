package extensions_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markus/claude-setup/internal/extensions"
)

func fakeClaude(t *testing.T) (path string, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "calls.txt")
	path = filepath.Join(dir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", argsFile)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return path, argsFile
}

func TestInstallerPlugin(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "superpowers", Source: "claude-plugins-official/superpowers"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install claude-plugins-official/superpowers") {
		t.Errorf("expected plugin install call, got: %q", string(data))
	}
}

func TestInstallerMCP(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "github", Scope: "local", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if !strings.Contains(got, "mcp add github --scope local npx -y @modelcontextprotocol/server-github") {
		t.Errorf("expected mcp add call, got: %q", got)
	}
}

func TestInstallerMCPDefaultScope(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "k8s", Command: "kubectl-mcp"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "--scope local") {
		t.Errorf("expected default scope local, got: %q", string(data))
	}
}

func TestInstallerEmpty(t *testing.T) {
	inst := extensions.Installer{Claude: "/nonexistent/claude"}
	if err := inst.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("Install empty: %v", err)
	}
}
