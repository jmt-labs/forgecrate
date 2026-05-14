package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Request struct {
	SourceDir string
	DestDir   string
	Profile   string
	Flavors   []string
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
	candidates := []string{
		filepath.Join(req.SourceDir, "base", filename),
		filepath.Join(req.SourceDir, "profiles", req.Profile, filename),
	}
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

func composeJSON(req Request) error {
	basePath := filepath.Join(req.SourceDir, "base", ".claude", "settings.json")
	data, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}
	merged := string(data)

	profilePath := filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "settings.json")
	if override, err := os.ReadFile(profilePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return err
		}
	}

	overridePath := filepath.Join(req.DestDir, ".claude", "overrides", "settings.override.json")
	if override, err := os.ReadFile(overridePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return err
		}
	}

	var v any
	if err := json.Unmarshal([]byte(merged), &v); err != nil {
		return fmt.Errorf("merged JSON invalid: %w", err)
	}

	dst := filepath.Join(req.DestDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(dst, []byte(merged), 0644)
}

func composeSkills(req Request) error {
	skillsDest := filepath.Join(req.DestDir, ".claude", "commands")
	if err := os.MkdirAll(skillsDest, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	srcDirs := []string{
		filepath.Join(req.SourceDir, "base", ".claude", "commands"),
		filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "commands"),
	}
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
