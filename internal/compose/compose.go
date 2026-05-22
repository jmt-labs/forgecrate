package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Request struct {
	SourceDir      string
	DestDir        string
	Profile        string
	Flavors        []string
	PermissionMode string
	SkipSettings   bool
}

func Run(req Request) error {
	if err := composeMarkdown(req, "CLAUDE.md"); err != nil {
		return fmt.Errorf("CLAUDE.md: %w", err)
	}
	if err := composeMarkdown(req, "AGENTS.md"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md: %w", err)
	}
	if err := composeJSON(req); err != nil {
		return fmt.Errorf("settings.json: %w", err)
	}
	if err := composeSkills(req); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
	return nil
}

// RunSingle führt compose für eine einzelne Markdown-Datei aus.
// Gibt os.ErrNotExist zurück wenn keine Layers vorhanden sind.
func RunSingle(req Request, filename string) error {
	return composeMarkdown(req, filename)
}

// RunCommands deployt Slash-Commands aus allen Layern nach DestDir.
func RunCommands(req Request) error {
	return composeSkills(req)
}

func composeMarkdown(req Request, filename string) error {
	layers := collectMarkdownLayers(req, filename)
	if len(layers) == 0 {
		return os.ErrNotExist
	}

	existing := ""
	if data, err := os.ReadFile(filepath.Join(req.DestDir, filename)); err == nil {
		existing = string(data)
	}

	result := MergeMarkdown(layers, existing)
	return os.WriteFile(filepath.Join(req.DestDir, filename), []byte(result), 0644)
}

func collectMarkdownLayers(req Request, filename string) []string {
	var layers []string
	profileCfg := LoadProfileConfig(req.SourceDir, req.Profile)

	candidates := []string{
		filepath.Join(req.SourceDir, "base", filename),
	}
	for _, parent := range profileCfg.Extends {
		candidates = append(candidates, filepath.Join(req.SourceDir, "profiles", parent, filename))
	}
	candidates = append(candidates, filepath.Join(req.SourceDir, "profiles", req.Profile, filename))
	for _, flavor := range req.Flavors {
		candidates = append(candidates, filepath.Join(req.SourceDir, "flavors", flavor, filename))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			layers = append(layers, string(data))
		}
	}
	return layers
}

// ComposeSettings berechnet den gemergten settings.json-Inhalt ohne ihn zu schreiben.
func ComposeSettings(req Request) ([]byte, error) {
	basePath := filepath.Join(req.SourceDir, "base", ".claude", "settings.json")
	data, err := os.ReadFile(basePath)
	if err != nil {
		return nil, err
	}
	merged := string(data)

	profilePath := filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "settings.json")
	if override, err := os.ReadFile(profilePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return nil, err
		}
	}

	overridePath := filepath.Join(req.DestDir, ".claude", "overrides", "settings.override.json")
	if override, err := os.ReadFile(overridePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return nil, err
		}
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(merged), &m); err != nil {
		return nil, fmt.Errorf("merged JSON invalid: %w", err)
	}
	if req.PermissionMode != "" {
		m["permissionMode"] = req.PermissionMode
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return append(out, '\n'), nil
}

func composeJSON(req Request) error {
	if req.SkipSettings {
		return nil
	}
	content, err := ComposeSettings(req)
	if err != nil {
		return err
	}
	dst := filepath.Join(req.DestDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(dst, content, 0644)
}

func composeSkills(req Request) error {
	skillsDest := filepath.Join(req.DestDir, ".claude", "commands")
	if err := os.MkdirAll(skillsDest, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	profileCfg := LoadProfileConfig(req.SourceDir, req.Profile)

	srcDirs := []string{
		filepath.Join(req.SourceDir, "base", ".claude", "commands"),
	}
	for _, parent := range profileCfg.Extends {
		srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "profiles", parent, ".claude", "commands"))
	}
	srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "commands"))
	for _, flavor := range req.Flavors {
		srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "flavors", flavor, ".claude", "commands"))
	}

	var existing []string
	for _, d := range srcDirs {
		if _, err := os.Stat(d); err == nil {
			existing = append(existing, d)
		}
	}

	return MergeSkills(existing, skillsDest)
}
