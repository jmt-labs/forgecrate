package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeYAML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".forgecrate.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestPromptSubmitOutput_BlockList(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `profile: backend
flavors:
  - tdd
  - strict-review
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: backend") {
		t.Errorf("expected 'Profil: backend', got: %s", out)
	}
	if !strings.Contains(out, "tdd") {
		t.Errorf("expected 'tdd' in output, got: %s", out)
	}
	if !strings.Contains(out, "strict-review") {
		t.Errorf("expected 'strict-review' in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_InlineList(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `profile: backend
flavors: [tdd, strict-review]
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: backend") {
		t.Errorf("expected 'Profil: backend', got: %s", out)
	}
	if !strings.Contains(out, "tdd") {
		t.Errorf("expected 'tdd' in output, got: %s", out)
	}
	if !strings.Contains(out, "strict-review") {
		t.Errorf("expected 'strict-review' in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_CommentIgnored(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `# flavors: [this-should-not-appear]
profile: backend
flavors:
  - tdd
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "this-should-not-appear") {
		t.Errorf("comment value must not appear in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_MissingFile(t *testing.T) {
	dir := t.TempDir()
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "unbekannt") {
		t.Errorf("expected 'unbekannt' for missing config, got: %s", out)
	}
	if !strings.Contains(out, "keine") {
		t.Errorf("expected 'keine' for missing config, got: %s", out)
	}
}

func TestPromptSubmitOutput_FallbackFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-setup.yaml")
	if err := os.WriteFile(path, []byte("profile: frontend\nflavors:\n  - github\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: frontend") {
		t.Errorf("expected 'Profil: frontend', got: %s", out)
	}
}

func TestPromptSubmitOutput_ContainsSkillsLine(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "profile: backend\nflavors:\n  - tdd\n")
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Pflicht-Skills") {
		t.Errorf("expected 'Pflicht-Skills' line, got: %s", out)
	}
	if !strings.Contains(out, "brainstorming") {
		t.Errorf("expected 'brainstorming' in output, got: %s", out)
	}
}
