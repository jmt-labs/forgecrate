package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("base/extensions.yaml", `
plugins:
  - name: superpowers
    source: superpowers
mcp:
  - name: fetch
    command: npx
    args: ["-y", "mcp-fetch-server"]
`)
	write("flavors/strict-review/extensions.yaml", `
plugins:
  - name: pr-review-toolkit
    source: pr-review-toolkit
`)
	return src
}

func fakeClaudeBin(t *testing.T) (bin, calls string) {
	t.Helper()
	dir := t.TempDir()
	bin = filepath.Join(dir, "claude")
	calls = filepath.Join(dir, "calls.txt")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"mcp\" ] && [ \"$2\" = \"get\" ]; then exit 1; fi\n" +
		"echo \"$@\" >> " + calls + "\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return bin, calls
}

func TestRunHostSetup_ProjectScopeWritesMCPJson(t *testing.T) {
	src := makeSource(t)
	bin, calls := fakeClaudeBin(t)
	target := t.TempDir()
	var out strings.Builder

	opts := hostSetupOpts{
		Scope:       "project",
		Yes:         true,
		SkipPrereqs: true,
		ClaudeBin:   bin,
		TargetDir:   target,
	}
	if err := runHostSetup(src, opts, &out); err != nil {
		t.Fatalf("runHostSetup: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, ".mcp.json")); err != nil {
		t.Errorf("expected .mcp.json in project scope: %v", err)
	}
	got, _ := os.ReadFile(calls)
	if !strings.Contains(string(got), "plugin install --scope project superpowers") {
		t.Errorf("expected project-scope plugin install, got:\n%s", got)
	}
	if !strings.Contains(string(got), "pr-review-toolkit") {
		t.Errorf("union should include flavor plugin, got:\n%s", got)
	}
}

func TestRunHostSetup_HostScopeUsesUserScope(t *testing.T) {
	src := makeSource(t)
	bin, calls := fakeClaudeBin(t)
	target := t.TempDir()
	var out strings.Builder

	opts := hostSetupOpts{
		Scope:       "host",
		Yes:         true,
		SkipPrereqs: true,
		ClaudeBin:   bin,
		TargetDir:   target,
	}
	if err := runHostSetup(src, opts, &out); err != nil {
		t.Fatalf("runHostSetup: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, ".mcp.json")); err == nil {
		t.Error("host scope must NOT write a repo .mcp.json")
	}
	got, _ := os.ReadFile(calls)
	s := string(got)
	if !strings.Contains(s, "plugin install --scope user superpowers") {
		t.Errorf("expected user-scope plugin install, got:\n%s", s)
	}
	if !strings.Contains(s, "mcp add --scope user fetch") {
		t.Errorf("expected user-scope mcp add, got:\n%s", s)
	}
}

func TestRunHostSetup_DryRunRunsNothing(t *testing.T) {
	src := makeSource(t)
	bin, calls := fakeClaudeBin(t)
	target := t.TempDir()
	var out strings.Builder

	opts := hostSetupOpts{
		Scope:       "host",
		Yes:         true,
		DryRun:      true,
		SkipPrereqs: true,
		ClaudeBin:   bin,
		TargetDir:   target,
	}
	if err := runHostSetup(src, opts, &out); err != nil {
		t.Fatalf("runHostSetup: %v", err)
	}
	if _, err := os.Stat(calls); err == nil {
		t.Error("dry-run must not invoke claude")
	}
}

func TestRunHostSetup_InvalidScope(t *testing.T) {
	src := makeSource(t)
	if err := runHostSetup(src, hostSetupOpts{Scope: "nonsense"}, &strings.Builder{}); err == nil {
		t.Fatal("expected error for invalid scope")
	}
}
