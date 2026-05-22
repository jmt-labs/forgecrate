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
	_ = os.WriteFile(fakeClaude, []byte(script), 0755)

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
