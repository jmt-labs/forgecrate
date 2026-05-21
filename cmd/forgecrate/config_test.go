package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func TestConfigInteractive_WritesUpdatedConfig(t *testing.T) {
	srcDir := t.TempDir()

	for _, p := range []string{"backend", "frontend", "fullstack"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "profiles", p), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tdd", "strict-review", "github"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "flavors", f), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
	}

	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		for _, p := range []string{"backend", "frontend", "fullstack"} {
			if !slices.Contains(profiles, p) {
				t.Errorf("profiles missing %q, got %v", p, profiles)
			}
		}
		return "frontend", []string{"strict-review"}, nil
	}

	got, err := configInteractive(srcDir, cfg, stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Profile != "frontend" {
		t.Errorf("returned profile = %q, want frontend", got.Profile)
	}
	if !slices.Equal(got.Flavors, []string{"strict-review"}) {
		t.Errorf("returned flavors = %v, want [strict-review]", got.Flavors)
	}
}

func TestConfigInteractive_EmptyProfiles(t *testing.T) {
	srcDir := t.TempDir()
	// Kein profiles/-Verzeichnis im srcDir

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		t.Fatal("prompt should not be called when profiles are empty")
		return "", nil, nil
	}

	_, err := configInteractive(srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error for empty profiles, got nil")
	}
}

func TestConfigInteractive_PromptError(t *testing.T) {
	srcDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(srcDir, "profiles", "backend"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "flavors", "tdd"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		return "", nil, fmt.Errorf("abgebrochen")
	}

	_, err := configInteractive(srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error from prompt, got nil")
	}
}
