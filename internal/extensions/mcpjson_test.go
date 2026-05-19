package extensions_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func TestWriteMCPJsonStdio(t *testing.T) {
	dst := t.TempDir()
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "context7", Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}},
		},
	}

	if err := extensions.WriteMCPJson(dst, ext); err != nil {
		t.Fatalf("WriteMCPJson: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	if err != nil {
		t.Fatalf(".mcp.json missing: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	servers, ok := result["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers missing or wrong type")
	}
	entry, ok := servers["context7"].(map[string]interface{})
	if !ok {
		t.Fatal("context7 entry missing")
	}
	if entry["command"] != "npx" {
		t.Errorf("expected command npx, got %v", entry["command"])
	}
}

func TestWriteMCPJsonHTTP(t *testing.T) {
	dst := t.TempDir()
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "github", Transport: "http", URL: "https://api.githubcopilot.com/mcp/"},
		},
	}

	if err := extensions.WriteMCPJson(dst, ext); err != nil {
		t.Fatalf("WriteMCPJson: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	servers := result["mcpServers"].(map[string]interface{})
	entry := servers["github"].(map[string]interface{})
	if entry["type"] != "http" {
		t.Errorf("expected type http, got %v", entry["type"])
	}
	if entry["url"] != "https://api.githubcopilot.com/mcp/" {
		t.Errorf("unexpected url: %v", entry["url"])
	}
}

func TestWriteMCPJsonMergesExisting(t *testing.T) {
	dst := t.TempDir()

	initial := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "existing", Command: "tool"},
		},
	}
	if err := extensions.WriteMCPJson(dst, initial); err != nil {
		t.Fatalf("first write: %v", err)
	}

	second := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "new", Command: "other"},
		},
	}
	if err := extensions.WriteMCPJson(dst, second); err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	servers := result["mcpServers"].(map[string]interface{})

	if _, ok := servers["existing"]; !ok {
		t.Error("existing entry was lost after second write")
	}
	if _, ok := servers["new"]; !ok {
		t.Error("new entry missing after second write")
	}
}

func TestWriteMCPJsonEmptyIsNoop(t *testing.T) {
	dst := t.TempDir()
	ext := extensions.Extensions{}

	if err := extensions.WriteMCPJson(dst, ext); err != nil {
		t.Fatalf("WriteMCPJson: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".mcp.json")); err == nil {
		t.Error("expected no .mcp.json for empty extensions")
	}
}
