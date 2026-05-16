package extensions_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/extensions"
)

func fakeClaude(t *testing.T) (path string, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "calls.txt")
	path = filepath.Join(dir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", argsFile)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return path, argsFile
}

func fakeClaudeWithOutput(t *testing.T, output string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho %q\nexit 1\n", output)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return path
}

func TestInstallerPlugin(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "superpowers", Source: "superpowers"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install --scope project superpowers") {
		t.Errorf("expected project-scoped plugin install call, got: %q", string(data))
	}
}

func TestInstallerEmpty(t *testing.T) {
	inst := extensions.Installer{Claude: "/nonexistent/claude"}
	if err := inst.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("Install empty: %v", err)
	}
}

func TestInstallerPluginNotFoundReturnsError(t *testing.T) {
	claude := fakeClaudeWithOutput(t, `Plugin "unknown-plugin" not found in any configured marketplace`)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "unknown-plugin", Source: "unknown-plugin"},
		},
	}

	if err := inst.Install(ext); err == nil {
		t.Fatal("expected error for plugin not found in marketplace, got nil")
	}
}

func TestInstallerPluginNotFoundErrorWrapsExitError(t *testing.T) {
	claude := fakeClaudeWithOutput(t, `Plugin "x" not found in any configured marketplace`)
	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "x", Source: "x"}},
	}
	err := inst.Install(ext)
	if err == nil {
		t.Fatal("expected error for plugin not found in marketplace, got nil")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("expected wrapped *exec.ExitError, got type %T: %v", err, err)
	}
}
