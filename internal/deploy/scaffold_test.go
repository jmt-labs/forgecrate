package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldMemoryBank_CreatesFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcMB := filepath.Join(src, "base", "memory-bank")
	if err := os.MkdirAll(srcMB, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcMB, "projectbrief.md"), []byte("# Project Brief\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcMB, "techContext.md"), []byte("# Tech Context\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("scaffoldMemoryBank: %v", err)
	}

	for _, name := range []string{"projectbrief.md", "techContext.md"} {
		path := filepath.Join(dst, "memory-bank", name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestScaffoldMemoryBank_DoesNotOverwriteExisting(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcMB := filepath.Join(src, "base", "memory-bank")
	if err := os.MkdirAll(srcMB, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcMB, "projectbrief.md"), []byte("template\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	dstMB := filepath.Join(dst, "memory-bank")
	if err := os.MkdirAll(dstMB, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstMB, "projectbrief.md"), []byte("custom content\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("scaffoldMemoryBank: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dstMB, "projectbrief.md"))
	if string(data) != "custom content\n" {
		t.Errorf("existing file must not be overwritten, got: %q", data)
	}
}

func TestScaffoldMemoryBank_NoSourceDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("expected no error when source missing, got: %v", err)
	}
}
