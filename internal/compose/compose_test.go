package compose_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/compose"
)

func TestCompose(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base Claude")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":["Bash"]}}`)
	writeFile(t, src, "base/.claude/commands/base-skill.md", "# Base Skill")
	writeFile(t, src, "profiles/backend/CLAUDE.md", "## Backend Profile")
	writeFile(t, src, "flavors/tdd/CLAUDE.md", "## TDD Flavor")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "backend",
		Flavors:   []string{"tdd"},
	}

	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md missing: %v", err)
	}
	for _, want := range []string{"# Base Claude", "## Backend Profile", "## TDD Flavor"} {
		if !strings.Contains(string(claudeMD), want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "settings.json")); err != nil {
		t.Errorf("settings.json missing: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "commands", "base-skill.md")); err != nil {
		t.Errorf("base-skill.md missing: %v", err)
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
