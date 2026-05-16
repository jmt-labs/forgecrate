package extensions

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type mcpEntry struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type mcpJson struct {
	MCPServers map[string]mcpEntry `json:"mcpServers"`
}

// WriteMCPJson schreibt .mcp.json in destDir aus den MCP-Einträgen von ext.
// Existierende Einträge werden erhalten und nur überschrieben wenn der Name schon vorhanden ist.
func WriteMCPJson(destDir string, ext Extensions) error {
	if len(ext.MCP) == 0 {
		return nil
	}

	path := filepath.Join(destDir, ".mcp.json")

	existing := mcpJson{MCPServers: map[string]mcpEntry{}}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &existing)
		if existing.MCPServers == nil {
			existing.MCPServers = map[string]mcpEntry{}
		}
	}

	for _, m := range ext.MCP {
		entry := mcpEntry{}
		if m.Transport == "http" {
			entry.Type = "http"
			entry.URL = m.URL
		} else {
			entry.Command = m.Command
			if len(m.Args) > 0 {
				entry.Args = m.Args
			}
			if len(m.Env) > 0 {
				entry.Env = m.Env
			}
		}
		existing.MCPServers[m.Name] = entry
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
