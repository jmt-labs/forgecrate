package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListProfiles(t *testing.T) {
	src := t.TempDir()
	for _, p := range []string{"backend", "frontend", "fullstack"} {
		if err := os.MkdirAll(filepath.Join(src, "profiles", p), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	profiles, err := listDirs(filepath.Join(src, "profiles"))
	if err != nil {
		t.Fatalf("listDirs: %v", err)
	}
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d: %v", len(profiles), profiles)
	}
	found := map[string]bool{}
	for _, p := range profiles {
		found[p] = true
	}
	for _, want := range []string{"backend", "frontend", "fullstack"} {
		if !found[want] {
			t.Errorf("missing profile: %s", want)
		}
	}
}

func TestListFlavors(t *testing.T) {
	src := t.TempDir()
	for _, f := range []string{"tdd", "github", "strict-review"} {
		if err := os.MkdirAll(filepath.Join(src, "flavors", f), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	flavors, err := listDirs(filepath.Join(src, "flavors"))
	if err != nil {
		t.Fatalf("listDirs: %v", err)
	}
	if len(flavors) != 3 {
		t.Errorf("expected 3 flavors, got %d: %v", len(flavors), flavors)
	}
}

func TestListDirsEmpty(t *testing.T) {
	src := t.TempDir()
	dirs, err := listDirs(filepath.Join(src, "nonexistent"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected empty, got %v", dirs)
	}
}
