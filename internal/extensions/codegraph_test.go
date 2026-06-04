package extensions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func TestInitCodegraph_BinaryMissing_WarnsNoError(t *testing.T) {
	dir := t.TempDir()
	destDir := t.TempDir()

	t.Setenv("PATH", dir)

	var buf strings.Builder
	err := extensions.InitCodegraph(destDir, "codegraph", &buf)
	if err != nil {
		t.Errorf("expected no error when binary missing, got: %v", err)
	}
	if !strings.Contains(buf.String(), "codegraph") {
		t.Errorf("expected warning mentioning codegraph, got: %q", buf.String())
	}
}

func TestInitCodegraph_AlreadyInitialized_Skips(t *testing.T) {
	binDir := t.TempDir()
	fakeBin := filepath.Join(binDir, "codegraph")
	script := "#!/bin/sh\necho \"$@\" >> " + filepath.Join(binDir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	destDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(destDir, ".codegraph"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	var buf strings.Builder
	err := extensions.InitCodegraph(destDir, fakeBin, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(binDir, "calls.txt"))
	if len(calls) > 0 {
		t.Errorf("expected no codegraph calls when .codegraph/ exists, got: %s", calls)
	}
	if !strings.Contains(buf.String(), "🔵") {
		t.Errorf("expected 🔵 skip log, got: %q", buf.String())
	}
}

func TestInitCodegraph_NotInitialized_RunsInit(t *testing.T) {
	binDir := t.TempDir()
	fakeBin := filepath.Join(binDir, "codegraph")
	script := "#!/bin/sh\necho \"$@\" >> " + filepath.Join(binDir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeBin, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	destDir := t.TempDir()

	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	var buf strings.Builder
	err := extensions.InitCodegraph(destDir, fakeBin, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(binDir, "calls.txt"))
	if !strings.Contains(string(calls), "init -i") {
		t.Errorf("expected 'init -i' call, got: %s", calls)
	}
}
