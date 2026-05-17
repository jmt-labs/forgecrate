package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

func TestDescribeProfile(t *testing.T) {
	src := t.TempDir()
	mustMkdir(t, filepath.Join(src, "profiles", "backend"))
	mustWriteFile(t, filepath.Join(src, "profiles", "backend", "CLAUDE.md"), []byte("## Backend-Profil\n\n- API-Design: REST-First\n"))

	out, err := describeEntry(src, "profile", "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "=== PROFILE: backend ===") {
		t.Errorf("output missing header: %q", out)
	}
	if !strings.Contains(out, "Backend-Profil") {
		t.Errorf("output missing profile content: %q", out)
	}
}

func TestDescribeFlavor(t *testing.T) {
	src := t.TempDir()
	mustMkdir(t, filepath.Join(src, "flavors", "tdd"))
	mustMkdir(t, filepath.Join(src, "flavors", "tdd", "skills", "test-coverage"))
	mustWriteFile(t, filepath.Join(src, "flavors", "tdd", "CLAUDE.md"), []byte("## TDD-Flavor\n\n- Test schreiben → ausführen\n"))
	mustWriteFile(t, filepath.Join(src, "flavors", "tdd", "skills", "test-coverage", "SKILL.md"), []byte("# Test Coverage\n"))

	out, err := describeEntry(src, "flavor", "tdd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "=== FLAVOR: tdd ===") {
		t.Errorf("output missing header: %q", out)
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
	mustMkdir(t, filepath.Join(src, "flavors", "minimal"))
	mustWriteFile(t, filepath.Join(src, "flavors", "minimal", "CLAUDE.md"), []byte("## Minimal-Flavor\n"))

	out, err := describeEntry(src, "flavor", "minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Minimal-Flavor") {
		t.Errorf("output missing content: %q", out)
	}
	if strings.Contains(out, "Skills:") {
		t.Errorf("output should not contain Skills section when no skills present: %q", out)
	}
}

func TestDescribeUnknownKind(t *testing.T) {
	src := t.TempDir()
	_, err := describeEntry(src, "unknown", "foo")
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
	if !strings.Contains(err.Error(), "erlaubt: profile, flavor") {
		t.Errorf("error message should list valid kinds: %q", err.Error())
	}
}

func TestDescribeNotFound(t *testing.T) {
	src := t.TempDir()
	mustMkdir(t, filepath.Join(src, "profiles"))
	_, err := describeEntry(src, "profile", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent profile")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error message should contain the missing name: %q", err.Error())
	}
}
