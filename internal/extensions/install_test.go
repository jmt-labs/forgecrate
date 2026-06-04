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

func TestInstallSkipsAlreadyInstalledPlugin(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	// plugin list gibt den Plugin-Namen aus -> Install-Command darf nicht aufgerufen werden
	script := `#!/bin/sh
if [ "$2" = "list" ]; then
  echo "already-installed"
  exit 0
fi
echo "$@" >> ` + filepath.Join(dir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "already-installed", Source: "https://github.com/foo/bar", Method: ""},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(dir, "calls.txt"))
	if strings.Contains(string(calls), "plugin install") {
		t.Errorf("expected install to be skipped, but got calls: %s", calls)
	}
}

func TestInstallRunsInstallWhenPluginNotListed(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	// plugin list gibt etwas anderes aus -> Install-Command muss aufgerufen werden
	script := `#!/bin/sh
if [ "$2" = "list" ]; then
  echo "some-other-plugin"
  exit 0
fi
echo "$@" >> ` + filepath.Join(dir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "new-plugin", Source: "https://github.com/foo/new-plugin", Method: ""},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(dir, "calls.txt"))
	if !strings.Contains(string(calls), "plugin install") {
		t.Errorf("expected install call, got: %s", calls)
	}
}

func TestInstallDoesNotSkipWhenNameIsSubstring(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	// plugin list enthält "superpowers-extra" — "superpowers" darf NICHT als installiert gelten
	script := `#!/bin/sh
if [ "$2" = "list" ]; then
  echo "superpowers-extra"
  exit 0
fi
echo "$@" >> ` + filepath.Join(dir, "calls.txt") + "\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "superpowers", Source: "https://github.com/foo/superpowers", Method: ""},
		},
	}

	installer := extensions.Installer{Claude: fakeClaude, Dir: dir}
	if err := installer.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	calls, _ := os.ReadFile(filepath.Join(dir, "calls.txt"))
	if !strings.Contains(string(calls), "plugin install") {
		t.Errorf("superpowers sollte installiert werden, da es nur als Substring (superpowers-extra) vorkam: %s", calls)
	}
}
