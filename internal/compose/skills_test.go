package compose_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestMergeSkills(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	_ = os.WriteFile(filepath.Join(src, "base.md"), []byte("# Base Skill"), 0644)
	_ = os.WriteFile(filepath.Join(src, "shared.md"), []byte("# Shared Base"), 0644)

	overridesDir := filepath.Join(dst, "overrides")
	_ = os.MkdirAll(overridesDir, 0755)
	_ = os.WriteFile(filepath.Join(overridesDir, "shared.md"), []byte("# My Override"), 0644)

	if err := compose.MergeSkills([]string{src}, dst); err != nil {
		t.Fatalf("MergeSkills: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dst, "base.md"))
	if err != nil {
		t.Fatalf("base.md not found: %v", err)
	}
	if string(content) != "# Base Skill" {
		t.Errorf("base.md: got %q", content)
	}

	override, _ := os.ReadFile(filepath.Join(overridesDir, "shared.md"))
	if string(override) != "# My Override" {
		t.Errorf("override was overwritten: got %q", override)
	}
}

func TestMergeSkillsLayerPrecedence(t *testing.T) {
	src1 := t.TempDir()
	src2 := t.TempDir()
	dst := t.TempDir()

	_ = os.WriteFile(filepath.Join(src1, "skill.md"), []byte("# Layer 1"), 0644)
	_ = os.WriteFile(filepath.Join(src2, "skill.md"), []byte("# Layer 2"), 0644)

	if err := compose.MergeSkills([]string{src1, src2}, dst); err != nil {
		t.Fatalf("MergeSkills: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dst, "skill.md"))
	if string(content) != "# Layer 2" {
		t.Errorf("layer 2 should win: got %q", string(content))
	}
}
