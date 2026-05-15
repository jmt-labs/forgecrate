package extensions

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

type MCP struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"`
	URL       string            `yaml:"url"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	Env       map[string]string `yaml:"env"`
}

type Extensions struct {
	Plugins []Plugin `yaml:"plugins"`
	MCP     []MCP    `yaml:"mcp"`
}

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
