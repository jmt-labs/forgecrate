package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

type profileCase struct {
	name           string
	contentMarker  string
	expectedSkills []string
}

type flavorCase struct {
	name             string
	contentMarker    string
	expectedSkills   []string
	expectedCommands []string
}

var allProfiles = []profileCase{
	{
		name:           "backend",
		contentMarker:  "Backend-Profil",
		expectedSkills: []string{"db-migration"},
	},
	{
		name:           "frontend",
		contentMarker:  "Frontend-Profil",
		expectedSkills: []string{"accessibility-audit", "ui-ux-audit"},
	},
	{
		name:          "fullstack",
		contentMarker: "Fullstack-Profil",
	},
}

var allFlavors = []flavorCase{
	{
		name:           "github",
		contentMarker:  "GitHub-Flavor",
		expectedSkills: []string{"github-release"},
	},
	{
		name:          "minimal",
		contentMarker: "Minimal-Flavor",
	},
	{
		name:           "strict-review",
		contentMarker:  "Strict-Review-Flavor",
		expectedSkills: []string{"pr-checklist"},
	},
	{
		name:           "tdd",
		contentMarker:  "TDD-Flavor",
		expectedSkills: []string{"test-coverage"},
	},
	{
		name:             "gitops",
		contentMarker:    "GitOps-Flavor",
		expectedSkills:   []string{"gitops-drift-check"},
		expectedCommands: []string{"forgecrate-gitops-status.md"},
	},
}

func TestAllProfiles(t *testing.T) {
	for _, tc := range allProfiles {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dst := t.TempDir()
			cfg := config.Config{
				Version: "1.0",
				Source:  "github.com/jmt-labs/forgecrate",
				Ref:     "main",
				Profile: tc.name,
				Flavors: []string{},
			}
			if err := deploy.Run(localSource(t), dst, cfg); err != nil {
				t.Fatalf("deploy.Run: %v", err)
			}

			claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
			if err != nil {
				t.Fatalf("CLAUDE.md not found: %v", err)
			}
			if !strings.Contains(string(claudeMD), tc.contentMarker) {
				t.Errorf("CLAUDE.md missing marker %q for profile %s", tc.contentMarker, tc.name)
			}

			for _, skill := range tc.expectedSkills {
				path := filepath.Join(dst, ".claude", "skills", skill, "SKILL.md")
				if _, err := os.Stat(path); err != nil {
					t.Errorf("profile skill missing: %s", skill)
				}
			}

			yamlContent, err := os.ReadFile(filepath.Join(dst, ".forgecrate.yaml"))
			if err != nil {
				t.Fatalf(".forgecrate.yaml not found: %v", err)
			}
			if !strings.Contains(string(yamlContent), tc.name) {
				t.Errorf(".forgecrate.yaml does not contain profile %q", tc.name)
			}
		})
	}
}

func TestAllFlavors(t *testing.T) {
	for _, tc := range allFlavors {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dst := t.TempDir()
			cfg := config.Config{
				Version: "1.0",
				Source:  "github.com/jmt-labs/forgecrate",
				Ref:     "main",
				Profile: "backend",
				Flavors: []string{tc.name},
			}
			if err := deploy.Run(localSource(t), dst, cfg); err != nil {
				t.Fatalf("deploy.Run: %v", err)
			}

			claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
			if err != nil {
				t.Fatalf("CLAUDE.md not found: %v", err)
			}
			if !strings.Contains(string(claudeMD), tc.contentMarker) {
				t.Errorf("CLAUDE.md missing marker %q for flavor %s", tc.contentMarker, tc.name)
			}

			for _, skill := range tc.expectedSkills {
				path := filepath.Join(dst, ".claude", "skills", skill, "SKILL.md")
				if _, err := os.Stat(path); err != nil {
					t.Errorf("flavor skill missing: %s", skill)
				}
			}

			for _, cmd := range tc.expectedCommands {
				path := filepath.Join(dst, ".claude", "commands", cmd)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("flavor command missing: %s", cmd)
				}
			}

			yamlContent, err := os.ReadFile(filepath.Join(dst, ".forgecrate.yaml"))
			if err != nil {
				t.Fatalf(".forgecrate.yaml not found: %v", err)
			}
			if !strings.Contains(string(yamlContent), tc.name) {
				t.Errorf(".forgecrate.yaml does not contain flavor %q", tc.name)
			}
		})
	}
}

func TestMultipleFlavors(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "github"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}
	content := string(claudeMD)
	for _, marker := range []string{"TDD-Flavor", "GitHub-Flavor"} {
		if !strings.Contains(content, marker) {
			t.Errorf("CLAUDE.md missing marker %q", marker)
		}
	}

	for _, skill := range []string{"test-coverage", "github-release"} {
		path := filepath.Join(dst, ".claude", "skills", skill, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("skill missing: %s", skill)
		}
	}
}
