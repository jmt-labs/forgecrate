package extensions

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Plugin describes a Claude Code plugin to install into the target repo.
type Plugin struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Method string `yaml:"method"`
}

// MCP describes an MCP server entry to be written into .mcp.json.
type MCP struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"`
	URL       string            `yaml:"url"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	Env       map[string]string `yaml:"env"`
}

// Extensions holds all plugin and MCP server definitions from one extensions.yaml layer.
type Extensions struct {
	Plugins []Plugin `yaml:"plugins"`
	MCP     []MCP    `yaml:"mcp"`
}

// Load parses an extensions.yaml file and returns the decoded Extensions.
func Load(path string) (Extensions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Extensions{}, err
	}
	var ext Extensions
	if err := yaml.Unmarshal(data, &ext); err != nil {
		return Extensions{}, err
	}
	return ext, nil
}

// Merge combines multiple Extensions layers; first occurrence of a name wins.
func Merge(layers []Extensions) Extensions {
	var result Extensions
	seenPlugin := map[string]bool{}
	seenMCP := map[string]bool{}

	for _, layer := range layers {
		for _, p := range layer.Plugins {
			if !seenPlugin[p.Name] {
				seenPlugin[p.Name] = true
				result.Plugins = append(result.Plugins, p)
			}
		}
		for _, m := range layer.MCP {
			if !seenMCP[m.Name] {
				seenMCP[m.Name] = true
				result.MCP = append(result.MCP, m)
			}
		}
	}
	return result
}
