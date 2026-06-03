package extensions_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func writeExt(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectAll_UnionAcrossLayers(t *testing.T) {
	src := t.TempDir()

	writeExt(t, filepath.Join(src, "base", "extensions.yaml"), `
plugins:
  - name: superpowers
    source: superpowers
mcp:
  - name: github
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
`)
	writeExt(t, filepath.Join(src, "profiles", "backend", "extensions.yaml"), `
plugins: []
mcp:
  - name: db
    command: npx
    args: ["-y", "db-mcp"]
`)
	writeExt(t, filepath.Join(src, "flavors", "strict-review", "extensions.yaml"), `
plugins:
  - name: pr-review-toolkit
    source: pr-review-toolkit
`)
	// Flavor with the legacy/typo key "mcpServers" must be ignored gracefully.
	writeExt(t, filepath.Join(src, "flavors", "github", "extensions.yaml"), `
plugins: []
mcpServers: []
`)

	got, err := extensions.CollectAll(src)
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}

	wantPlugins := map[string]bool{"superpowers": true, "pr-review-toolkit": true}
	if len(got.Plugins) != len(wantPlugins) {
		t.Fatalf("plugins = %v, want names %v", got.Plugins, wantPlugins)
	}
	for _, p := range got.Plugins {
		if !wantPlugins[p.Name] {
			t.Errorf("unexpected plugin %q", p.Name)
		}
	}

	wantMCP := map[string]bool{"github": true, "db": true}
	if len(got.MCP) != len(wantMCP) {
		t.Fatalf("mcp = %v, want names %v", got.MCP, wantMCP)
	}
	for _, m := range got.MCP {
		if !wantMCP[m.Name] {
			t.Errorf("unexpected mcp %q", m.Name)
		}
	}
}

func TestCollectAll_BaseWinsOnConflict(t *testing.T) {
	src := t.TempDir()

	writeExt(t, filepath.Join(src, "base", "extensions.yaml"), `
mcp:
  - name: github
    command: npx
    args: ["-y", "canonical-github"]
`)
	writeExt(t, filepath.Join(src, "flavors", "x", "extensions.yaml"), `
mcp:
  - name: github
    command: npx
    args: ["-y", "override-github"]
`)

	got, err := extensions.CollectAll(src)
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}
	if len(got.MCP) != 1 {
		t.Fatalf("want 1 mcp, got %v", got.MCP)
	}
	if got.MCP[0].Args[1] != "canonical-github" {
		t.Errorf("base should win, got args %v", got.MCP[0].Args)
	}
}

func TestCollectAll_MissingDirsAreSkipped(t *testing.T) {
	src := t.TempDir()
	writeExt(t, filepath.Join(src, "base", "extensions.yaml"), `
plugins:
  - name: only-base
    source: only-base
`)
	// No profiles/ or flavors/ dirs at all.
	got, err := extensions.CollectAll(src)
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}
	if len(got.Plugins) != 1 || got.Plugins[0].Name != "only-base" {
		t.Errorf("got %v, want only-base", got.Plugins)
	}
}
