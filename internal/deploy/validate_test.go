package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func catalogFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, p := range []string{"backend", "frontend", "fullstack"} {
		if err := os.MkdirAll(filepath.Join(dir, "profiles", p), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tdd", "minimal", "strict-review"} {
		if err := os.MkdirAll(filepath.Join(dir, "flavors", f), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestValidateSelection_Valid(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: "backend", Flavors: []string{"tdd", "minimal"}}
	if err := validateSelection(src, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSelection_UnknownProfileSuggests(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: "backendx"}
	err := validateSelection(src, cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `meintest du "backend"`) {
		t.Errorf("expected suggestion, got: %v", err)
	}
}

func TestValidateSelection_UnknownProfileNoSuggestionListsAvailable(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: "zzzzzz"}
	err := validateSelection(src, cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "meintest du") {
		t.Errorf("did not expect suggestion: %v", err)
	}
	for _, p := range []string{"backend", "frontend", "fullstack"} {
		if !strings.Contains(err.Error(), p) {
			t.Errorf("expected %q in available list: %v", p, err)
		}
	}
}

func TestValidateSelection_EmptyProfile(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: ""}
	err := validateSelection(src, cfg)
	if err == nil || !strings.Contains(err.Error(), "kein Profil angegeben") {
		t.Fatalf("expected 'kein Profil angegeben', got: %v", err)
	}
}

func TestValidateSelection_UnknownFlavorSuggests(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: "backend", Flavors: []string{"tddx"}}
	err := validateSelection(src, cfg)
	if err == nil || !strings.Contains(err.Error(), `meintest du "tdd"`) {
		t.Fatalf("expected suggestion for tdd, got: %v", err)
	}
}

func TestValidateSelection_MixedFlavors(t *testing.T) {
	src := catalogFixture(t)
	cfg := config.Config{Profile: "backend", Flavors: []string{"tdd", "nope"}}
	if err := validateSelection(src, cfg); err == nil {
		t.Fatal("expected error for unknown flavor")
	}
}

func TestValidateSelection_EmptyCatalogSkips(t *testing.T) {
	src := t.TempDir() // no profiles/ or flavors/ → nothing to validate against
	cfg := config.Config{Profile: "backend", Flavors: []string{"whatever"}}
	if err := validateSelection(src, cfg); err != nil {
		t.Fatalf("empty catalog should skip validation, got: %v", err)
	}
}

func TestRunRejectsUnknownFlavorWithoutWriting(t *testing.T) {
	src := catalogFixture(t)
	// Minimaler base-Layer: ohne Validierung würde Run anschließend schreiben.
	if err := os.MkdirAll(filepath.Join(src, "base", ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "base", "CLAUDE.md"), []byte("# Base"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "base", ".claude", "settings.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	cfg := config.Config{Profile: "backend", Flavors: []string{"tddx"}}
	if err := Run(src, dst, cfg); err == nil {
		t.Fatal("expected Run to reject unknown flavor")
	}
	if _, err := os.Stat(filepath.Join(dst, ".forgecrate.yaml")); !os.IsNotExist(err) {
		t.Errorf(".forgecrate.yaml should not be written on validation failure")
	}
	if _, err := os.Stat(filepath.Join(dst, ".claude")); !os.IsNotExist(err) {
		t.Errorf(".claude should not be written on validation failure")
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"tdd", "tddd", 1},
		{"abc", "xyz", 3},
		{"backend", "backend", 0},
		{"kitten", "sitting", 3},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}
