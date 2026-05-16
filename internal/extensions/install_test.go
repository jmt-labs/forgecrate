package extensions_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/extensions"
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

// fakeClaudeWithOutput creates a fake claude binary that outputs a fixed message and exits with code 1.
func fakeClaudeWithOutput(t *testing.T, output string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho %q\nexit 1\n", output)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return path
}

func TestInstallerPlugin(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "superpowers", Source: "superpowers"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install --scope project superpowers") {
		t.Errorf("expected project-scoped plugin install call, got: %q", string(data))
	}
}

func TestInstallerMCP(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "github", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if !strings.Contains(got, "mcp add github --scope project npx -- -y @modelcontextprotocol/server-github") {
		t.Errorf("expected mcp add call with -- separator, got: %q", got)
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
	if !strings.Contains(string(data), "--scope project") {
		t.Errorf("expected default scope project, got: %q", string(data))
	}
}

func TestInstallerMCPHTTP(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{
				Name:      "github",
				Transport: "http",
				URL:       "https://api.githubcopilot.com/mcp/",
			},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if !strings.Contains(got, "mcp add --transport http github https://api.githubcopilot.com/mcp/ --scope project") {
		t.Errorf("expected http mcp add call, got: %q", got)
	}
}

func TestInstallerMCPHTTPEnv(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{
				Name:      "github",
				Transport: "http",
				URL:       "https://api.githubcopilot.com/mcp/",
				Env:       map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": "tok"},
			},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "mcp add --transport http") {
		t.Errorf("expected http transport call, got: %q", string(data))
	}
}

func TestInstallerEmpty(t *testing.T) {
	inst := extensions.Installer{Claude: "/nonexistent/claude"}
	if err := inst.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("Install empty: %v", err)
	}
}

func TestInstallerMCPWithoutArgs(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "mytool", Command: "mytool-bin"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if strings.Contains(got, " -- ") {
		t.Errorf("no standalone -- separator expected when no args: %q", got)
	}
	if !strings.Contains(got, "mcp add mytool --scope project mytool-bin") {
		t.Errorf("unexpected call: %q", got)
	}
}

func TestInstallerMCPAlreadyExistsNoWarn(t *testing.T) {
	claude := fakeClaudeWithOutput(t, "MCP server github already exists in local config")

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "github", Transport: "http", URL: "https://api.githubcopilot.com/mcp/"},
		},
	}

	// Must not return an error — "already exists" is handled silently.
	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}
}

func TestInstallerPluginNotFoundNoWarn(t *testing.T) {
	claude := fakeClaudeWithOutput(t, `Plugin "unknown-plugin" not found in any configured marketplace`)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "unknown-plugin", Source: "unknown-plugin"},
		},
	}

	// Must not return an error — marketplace misses are handled silently.
	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}
}
