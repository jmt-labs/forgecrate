package extensions_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func TestCheckNodeVersion_SufficientVersion(t *testing.T) {
	dir := t.TempDir()
	fakeNode := filepath.Join(dir, "node")
	script := "#!/bin/sh\necho 'v22.0.0'\n"
	if err := os.WriteFile(fakeNode, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	if err := extensions.CheckNodeVersion(18); err != nil {
		t.Errorf("expected no error for v22, got: %v", err)
	}
}

func TestCheckNodeVersion_InsufficientVersion(t *testing.T) {
	dir := t.TempDir()
	fakeNode := filepath.Join(dir, "node")
	script := "#!/bin/sh\necho 'v16.0.0'\n"
	if err := os.WriteFile(fakeNode, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	err := extensions.CheckNodeVersion(18)
	if err == nil {
		t.Fatal("expected error for v16, got nil")
	}
}

func TestCheckNodeVersion_MissingBinary(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	err := extensions.CheckNodeVersion(18)
	if err == nil {
		t.Fatal("expected error when node is missing, got nil")
	}
}
