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

func TestInstallPluginNotFoundSkipped(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	script := "#!/bin/sh\necho 'not found in any configured marketplace'\nexit 1\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var buf strings.Builder
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "missing", Source: "org/missing", Method: "marketplace"},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir, Out: &buf}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("expected no error (plugin skip), got: %v", err)
	}
	if !strings.Contains(buf.String(), "missing") {
		t.Errorf("expected warning mentioning plugin name, got: %s", buf.String())
	}
}

func TestInstallEmptyPluginsSucceeds(t *testing.T) {
	installer := extensions.Installer{Claude: "echo"}
	if err := installer.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
