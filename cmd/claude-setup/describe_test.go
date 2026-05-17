package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDescribeProfile(t *testing.T) {
	src := t.TempDir()
	content := "## Backend-Profil\n\n- API-Design: REST-First\n"
	os.MkdirAll(filepath.Join(src, "profiles", "backend"), 0755)
	os.WriteFile(filepath.Join(src, "profiles", "backend", "CLAUDE.md"), []byte(content), 0644)

	out, err := describeEntry(src, "profile", "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Backend-Profil") {
		t.Errorf("output missing profile content: %q", out)
	}
}

func TestDescribeFlavor(t *testing.T) {
	src := t.TempDir()
	content := "## TDD-Flavor\n\n- Test schreiben → ausführen\n"
	os.MkdirAll(filepath.Join(src, "flavors", "tdd"), 0755)
	os.MkdirAll(filepath.Join(src, "flavors", "tdd", "skills", "test-coverage"), 0755)
	os.WriteFile(filepath.Join(src, "flavors", "tdd", "CLAUDE.md"), []byte(content), 0644)
	os.WriteFile(filepath.Join(src, "flavors", "tdd", "skills", "test-coverage", "SKILL.md"), []byte("# Test Coverage\n"), 0644)

	out, err := describeEntry(src, "flavor", "tdd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "TDD-Flavor") {
		t.Errorf("output missing flavor content: %q", out)
	}
	if !strings.Contains(out, "test-coverage") {
		t.Errorf("output missing skill name: %q", out)
	}
}

func TestDescribeFlavorNoSkills(t *testing.T) {
	src := t.TempDir()
	os.MkdirAll(filepath.Join(src, "flavors", "minimal"), 0755)
	os.WriteFile(filepath.Join(src, "flavors", "minimal", "CLAUDE.md"), []byte("## Minimal-Flavor\n"), 0644)

	out, err := describeEntry(src, "flavor", "minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Minimal-Flavor") {
		t.Errorf("output missing content: %q", out)
	}
}

func TestDescribeUnknownKind(t *testing.T) {
	src := t.TempDir()
	_, err := describeEntry(src, "unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown kind")
	}
}

func TestDescribeNotFound(t *testing.T) {
	src := t.TempDir()
	os.MkdirAll(filepath.Join(src, "profiles"), 0755)
	_, err := describeEntry(src, "profile", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}
