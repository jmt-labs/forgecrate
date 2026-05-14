package extensions_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/markus/claude-setup/internal/extensions"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	content := `
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers
mcp:
  - name: github
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      TOKEN: abc
`
	os.WriteFile(filepath.Join(dir, "extensions.yaml"), []byte(content), 0644)

	ext, err := extensions.Load(filepath.Join(dir, "extensions.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ext.Plugins) != 1 || ext.Plugins[0].Name != "superpowers" {
		t.Errorf("plugins: %+v", ext.Plugins)
	}
	if len(ext.MCP) != 1 || ext.MCP[0].Name != "github" {
		t.Errorf("mcp: %+v", ext.MCP)
	}
	if ext.MCP[0].Env["TOKEN"] != "abc" {
		t.Errorf("env: %+v", ext.MCP[0].Env)
	}
}

func TestLoadHTTPTransport(t *testing.T) {
	dir := t.TempDir()
	content := `
mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/
    scope: local
`
	os.WriteFile(filepath.Join(dir, "extensions.yaml"), []byte(content), 0644)

	ext, err := extensions.Load(filepath.Join(dir, "extensions.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ext.MCP) != 1 {
		t.Fatalf("expected 1 MCP, got %d", len(ext.MCP))
	}
	if ext.MCP[0].Transport != "http" {
		t.Errorf("Transport: got %q, want %q", ext.MCP[0].Transport, "http")
	}
	if ext.MCP[0].URL != "https://api.githubcopilot.com/mcp/" {
		t.Errorf("URL: got %q", ext.MCP[0].URL)
	}
}

func TestLoadNotExist(t *testing.T) {
	_, err := extensions.Load("/nonexistent/extensions.yaml")
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist, got: %v", err)
	}
}

func TestMergeFirstWins(t *testing.T) {
	base := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "superpowers", Source: "source-a"}},
	}
	flavor := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "superpowers", Source: "source-b"}},
		MCP:     []extensions.MCP{{Name: "github", Command: "npx"}},
	}

	merged := extensions.Merge([]extensions.Extensions{base, flavor})

	if len(merged.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(merged.Plugins))
	}
	if merged.Plugins[0].Source != "source-a" {
		t.Errorf("first-wins failed: got source %q", merged.Plugins[0].Source)
	}
	if len(merged.MCP) != 1 || merged.MCP[0].Name != "github" {
		t.Errorf("mcp: %+v", merged.MCP)
	}
}

func TestMergeEmpty(t *testing.T) {
	merged := extensions.Merge(nil)
	if len(merged.Plugins) != 0 || len(merged.MCP) != 0 {
		t.Errorf("expected empty, got: %+v", merged)
	}
}
