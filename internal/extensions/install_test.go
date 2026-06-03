package extensions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func TestInstallUsesMarketplaceForMethodMarketplace(t *testing.T) {
	// Fake claude-binary schreiben
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	script := "#!/bin/sh\necho \"$@\" >> " + filepath.Join(dir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "marketplace-plugin", Source: "org/plugin", Method: "marketplace"},
			{Name: "direct-plugin", Source: "https://github.com/foo/bar", Method: ""},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(dir, "calls.txt"))
	callsStr := string(calls)

	if !strings.Contains(callsStr, "plugin marketplace add org/plugin") {
		t.Errorf("expected marketplace call, got:\n%s", callsStr)
	}
	if !strings.Contains(callsStr, "plugin install --scope project https://github.com/foo/bar") {
		t.Errorf("expected install call, got:\n%s", callsStr)
	}
}

func TestInstallPluginNotFoundReturnsError(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	script := "#!/bin/sh\necho 'not found in any configured marketplace'\nexit 1\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "missing", Source: "org/missing", Method: "marketplace"},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir}
	err := installer.Install(ext)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found in marketplace") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInstallEmptyPluginsSucceeds(t *testing.T) {
	installer := extensions.Installer{Claude: "echo"}
	if err := installer.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func recordingClaude(t *testing.T) (claudeBin, callsFile string) {
	t.Helper()
	dir := t.TempDir()
	claudeBin = filepath.Join(dir, "claude")
	callsFile = filepath.Join(dir, "calls.txt")
	script := "#!/bin/sh\necho \"$@\" >> " + callsFile + "\n"
	if err := os.WriteFile(claudeBin, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return claudeBin, callsFile
}

func TestInstallUserScope(t *testing.T) {
	claudeBin, calls := recordingClaude(t)
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "sp", Source: "superpowers"}},
	}
	installer := extensions.Installer{Claude: claudeBin, Scope: "user"}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}
	got, _ := os.ReadFile(calls)
	if !strings.Contains(string(got), "plugin install --scope user superpowers") {
		t.Errorf("expected user scope, got:\n%s", got)
	}
}

func TestInstallDefaultScopeIsProject(t *testing.T) {
	claudeBin, calls := recordingClaude(t)
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "sp", Source: "superpowers"}},
	}
	installer := extensions.Installer{Claude: claudeBin} // empty Scope
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}
	got, _ := os.ReadFile(calls)
	if !strings.Contains(string(got), "plugin install --scope project superpowers") {
		t.Errorf("expected project scope by default, got:\n%s", got)
	}
}

func TestInstallDryRunRunsNothing(t *testing.T) {
	claudeBin, calls := recordingClaude(t)
	var out strings.Builder
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "sp", Source: "superpowers"}},
	}
	installer := extensions.Installer{Claude: claudeBin, DryRun: true, Out: &out}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(calls); err == nil {
		t.Error("dry-run must not invoke claude")
	}
	if !strings.Contains(out.String(), "superpowers") {
		t.Errorf("dry-run should print the planned action, got:\n%s", out.String())
	}
}

func TestInstallMCPUser_StdioAndHTTP(t *testing.T) {
	// Fake claude where `mcp get` always fails (server not yet registered),
	// so every add is attempted.
	dir := t.TempDir()
	claudeBin := filepath.Join(dir, "claude")
	calls := filepath.Join(dir, "calls.txt")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"mcp\" ] && [ \"$2\" = \"get\" ]; then exit 1; fi\n" +
		"echo \"$@\" >> " + calls + "\n"
	if err := os.WriteFile(claudeBin, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "fetch", Command: "npx", Args: []string{"-y", "mcp-fetch-server"}},
			{Name: "remote", Transport: "http", URL: "https://example.com/mcp"},
			{Name: "withenv", Command: "npx", Args: []string{"-y", "x"}, Env: map[string]string{"K": "v"}},
		},
	}
	installer := extensions.Installer{Claude: claudeBin}
	if err := installer.InstallMCPUser(ext); err != nil {
		t.Fatalf("InstallMCPUser: %v", err)
	}
	got, _ := os.ReadFile(calls)
	s := string(got)
	if !strings.Contains(s, "mcp add --scope user fetch") || !strings.Contains(s, "-- npx -y mcp-fetch-server") {
		t.Errorf("stdio add missing/wrong, got:\n%s", s)
	}
	if !strings.Contains(s, "mcp add --scope user --transport http remote https://example.com/mcp") {
		t.Errorf("http add missing/wrong, got:\n%s", s)
	}
	if !strings.Contains(s, "--env K=v") {
		t.Errorf("env flag missing, got:\n%s", s)
	}
}

func TestInstallMCPUser_SkipsExisting(t *testing.T) {
	// Fake claude where `mcp get` succeeds (already registered) -> add skipped.
	dir := t.TempDir()
	claudeBin := filepath.Join(dir, "claude")
	calls := filepath.Join(dir, "calls.txt")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"mcp\" ] && [ \"$2\" = \"get\" ]; then exit 0; fi\n" +
		"echo \"$@\" >> " + calls + "\n"
	if err := os.WriteFile(claudeBin, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		MCP: []extensions.MCP{{Name: "fetch", Command: "npx", Args: []string{"-y", "x"}}},
	}
	installer := extensions.Installer{Claude: claudeBin}
	if err := installer.InstallMCPUser(ext); err != nil {
		t.Fatalf("InstallMCPUser: %v", err)
	}
	if _, err := os.Stat(calls); err == nil {
		t.Error("add must be skipped when mcp get succeeds")
	}
}
