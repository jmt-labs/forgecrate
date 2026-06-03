package prereqs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/prereqs"
)

func fakeBin(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestDetect_PartitionsPresentAndMissing(t *testing.T) {
	binDir := t.TempDir()
	fakeBin(t, binDir, "claude")
	fakeBin(t, binDir, "node")
	// npx and codegraph intentionally absent
	t.Setenv("PATH", binDir)

	c := prereqs.Checker{}
	present, missing := c.Detect(prereqs.DefaultTools())

	has := func(list []prereqs.Tool, name string) bool {
		for _, tl := range list {
			if tl.Name == name {
				return true
			}
		}
		return false
	}

	if !has(present, "claude") || !has(present, "node") {
		t.Errorf("claude/node should be present, got present=%v", present)
	}
	if !has(missing, "npx") || !has(missing, "codegraph") {
		t.Errorf("npx/codegraph should be missing, got missing=%v", missing)
	}
}

func TestEnsureCodegraph_DryRunDoesNotExecute(t *testing.T) {
	binDir := t.TempDir()
	t.Setenv("PATH", binDir) // codegraph absent
	var out strings.Builder

	c := prereqs.Checker{Out: &out, DryRun: true, Assume: true}
	if err := c.EnsureCodegraph(); err != nil {
		t.Fatalf("EnsureCodegraph: %v", err)
	}
	if !strings.Contains(out.String(), "install.sh") {
		t.Errorf("dry-run should print the planned install command, got:\n%s", out.String())
	}
}

func TestEnsureCodegraph_AlreadyPresentIsNoop(t *testing.T) {
	binDir := t.TempDir()
	fakeBin(t, binDir, "codegraph")
	t.Setenv("PATH", binDir)
	var out strings.Builder

	c := prereqs.Checker{Out: &out, Assume: true}
	if err := c.EnsureCodegraph(); err != nil {
		t.Fatalf("EnsureCodegraph: %v", err)
	}
	if strings.Contains(out.String(), "install.sh") {
		t.Errorf("should not attempt install when codegraph present, got:\n%s", out.String())
	}
}
