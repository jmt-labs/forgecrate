package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func TestConfigInteractive_WritesUpdatedConfig(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, ".forgecrate.yaml"), []byte(
		"version: \"1.0\"\nsource: github.com/jmt-labs/forgecrate\nref: main\nprofile: backend\nflavors:\n  - tdd\n",
	), 0644); err != nil {
		t.Fatal(err)
	}
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
			found := false
			for _, got := range profiles {
				if got == p {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("profiles missing %q, got %v", p, profiles)
			}
		}
		return "frontend", []string{"strict-review"}, nil
	}

	got, err := configInteractive(dir, srcDir, cfg, stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Profile != "frontend" {
		t.Errorf("returned profile = %q, want frontend", got.Profile)
	}
	if len(got.Flavors) != 1 || got.Flavors[0] != "strict-review" {
		t.Errorf("returned flavors = %v, want [strict-review]", got.Flavors)
	}

	written, err := config.Read(filepath.Join(dir, ".forgecrate.yaml"))
	if err != nil {
		t.Fatalf("read back config: %v", err)
	}
	if written.Profile != "frontend" {
		t.Errorf("written profile = %q, want frontend", written.Profile)
	}
	if len(written.Flavors) != 1 || written.Flavors[0] != "strict-review" {
		t.Errorf("written flavors = %v, want [strict-review]", written.Flavors)
	}
}

func TestConfigInteractive_EmptyProfiles(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	// Kein profiles/-Verzeichnis im srcDir

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		t.Fatal("prompt should not be called when profiles are empty")
		return "", nil, nil
	}

	_, err := configInteractive(dir, srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error for empty profiles, got nil")
	}
}

func TestConfigInteractive_PromptError(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()

	for _, p := range []string{"backend"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "profiles", p), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tdd"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "flavors", f), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		return "", nil, fmt.Errorf("abgebrochen")
	}

	_, err := configInteractive(dir, srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error from prompt, got nil")
	}
}
