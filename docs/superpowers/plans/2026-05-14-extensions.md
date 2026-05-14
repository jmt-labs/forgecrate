# Extensions: Plugin & MCP Auto-Install — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Jeder Layer kann eine `extensions.yaml` deklarieren; `claude-setup init/update` merged alle aktiven Layer und installiert Plugins und MCP-Server via `claude`-CLI.

**Architecture:** Neues Package `internal/extensions` mit Typen, Merge-Logik und Installer. `deploy.Run()` ruft nach `copyHooks()` den neuen Schritt `installExtensions()` auf. Jeder Layer (base, profile, flavors) kann optional eine `extensions.yaml` enthalten.

**Tech Stack:** Go 1.24, `gopkg.in/yaml.v3` (bereits im go.mod), `os/exec`

---

## Dateistruktur

| Datei | Aktion | Zweck |
|---|---|---|
| `internal/extensions/extensions.go` | Neu | Typen, Load, Merge |
| `internal/extensions/extensions_test.go` | Neu | Tests für Load und Merge |
| `internal/extensions/install.go` | Neu | Installer (Shell-out zu `claude`) |
| `internal/extensions/install_test.go` | Neu | Tests mit Fake-Binary |
| `internal/deploy/deploy.go` | Ändern | `installExtensions()` nach `copyHooks()` |
| `internal/deploy/deploy_test.go` | Ändern | Test für Extensions-Integration |
| `base/extensions.yaml` | Neu | superpowers deklarieren |

---

### Task 1: Typen, Load und Merge

**Files:**
- Create: `internal/extensions/extensions.go`
- Create: `internal/extensions/extensions_test.go`

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/extensions/extensions_test.go
package extensions_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/markus/claude-setup/internal/extensions"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	content := `
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers
mcp:
  - name: github
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      TOKEN: abc
`
	os.WriteFile(filepath.Join(dir, "extensions.yaml"), []byte(content), 0644)

	ext, err := extensions.Load(filepath.Join(dir, "extensions.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ext.Plugins) != 1 || ext.Plugins[0].Name != "superpowers" {
		t.Errorf("plugins: %+v", ext.Plugins)
	}
	if len(ext.MCP) != 1 || ext.MCP[0].Name != "github" {
		t.Errorf("mcp: %+v", ext.MCP)
	}
	if ext.MCP[0].Env["TOKEN"] != "abc" {
		t.Errorf("env: %+v", ext.MCP[0].Env)
	}
}

func TestLoadNotExist(t *testing.T) {
	_, err := extensions.Load("/nonexistent/extensions.yaml")
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist, got: %v", err)
	}
}

func TestMergeFirstWins(t *testing.T) {
	base := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "superpowers", Source: "source-a"}},
	}
	flavor := extensions.Extensions{
		Plugins: []extensions.Plugin{{Name: "superpowers", Source: "source-b"}},
		MCP:     []extensions.MCP{{Name: "github", Command: "npx"}},
	}

	merged := extensions.Merge([]extensions.Extensions{base, flavor})

	if len(merged.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(merged.Plugins))
	}
	if merged.Plugins[0].Source != "source-a" {
		t.Errorf("first-wins failed: got source %q", merged.Plugins[0].Source)
	}
	if len(merged.MCP) != 1 || merged.MCP[0].Name != "github" {
		t.Errorf("mcp: %+v", merged.MCP)
	}
}

func TestMergeEmpty(t *testing.T) {
	merged := extensions.Merge(nil)
	if len(merged.Plugins) != 0 || len(merged.MCP) != 0 {
		t.Errorf("expected empty, got: %+v", merged)
	}
}
```

- [ ] **Schritt 2: Tests ausführen — müssen FAIL**

```
go test ./internal/extensions/...
```

Erwartet: `cannot find package`

- [ ] **Schritt 3: Implementierung schreiben**

```go
// internal/extensions/extensions.go
package extensions

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

type MCP struct {
	Name    string            `yaml:"name"`
	Scope   string            `yaml:"scope"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

type Extensions struct {
	Plugins []Plugin `yaml:"plugins"`
	MCP     []MCP    `yaml:"mcp"`
}

func Load(path string) (Extensions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Extensions{}, err
	}
	var ext Extensions
	if err := yaml.Unmarshal(data, &ext); err != nil {
		return Extensions{}, err
	}
	return ext, nil
}

func Merge(layers []Extensions) Extensions {
	var result Extensions
	seenPlugin := map[string]bool{}
	seenMCP := map[string]bool{}

	for _, layer := range layers {
		for _, p := range layer.Plugins {
			if !seenPlugin[p.Name] {
				seenPlugin[p.Name] = true
				result.Plugins = append(result.Plugins, p)
			}
		}
		for _, m := range layer.MCP {
			if !seenMCP[m.Name] {
				seenMCP[m.Name] = true
				result.MCP = append(result.MCP, m)
			}
		}
	}
	return result
}
```

- [ ] **Schritt 4: Tests ausführen — müssen PASS**

```
go test ./internal/extensions/...
```

Erwartet: `ok  github.com/markus/claude-setup/internal/extensions`

- [ ] **Schritt 5: Commit**

```bash
git add internal/extensions/extensions.go internal/extensions/extensions_test.go
git commit -m "feat(extensions): add types, Load and Merge"
```

---

### Task 2: Installer

**Files:**
- Create: `internal/extensions/install.go`
- Create: `internal/extensions/install_test.go`

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/extensions/install_test.go
package extensions_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markus/claude-setup/internal/extensions"
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

func TestInstallerPlugin(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "superpowers", Source: "claude-plugins-official/superpowers"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install claude-plugins-official/superpowers") {
		t.Errorf("expected plugin install call, got: %q", string(data))
	}
}

func TestInstallerMCP(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "github", Scope: "local", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if !strings.Contains(got, "mcp add github --scope local npx -y @modelcontextprotocol/server-github") {
		t.Errorf("expected mcp add call, got: %q", got)
	}
}

func TestInstallerMCPDefaultScope(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{Name: "k8s", Command: "kubectl-mcp"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "--scope local") {
		t.Errorf("expected default scope local, got: %q", string(data))
	}
}

func TestInstallerEmpty(t *testing.T) {
	inst := extensions.Installer{Claude: "/nonexistent/claude"}
	if err := inst.Install(extensions.Extensions{}); err != nil {
		t.Fatalf("Install empty: %v", err)
	}
}
```

- [ ] **Schritt 2: Tests ausführen — müssen FAIL**

```
go test ./internal/extensions/...
```

Erwartet: `undefined: extensions.Installer`

- [ ] **Schritt 3: Implementierung schreiben**

```go
// internal/extensions/install.go
package extensions

import (
	"log"
	"os"
	"os/exec"
)

type Installer struct {
	Claude string
}

func NewInstaller() Installer {
	return Installer{Claude: "claude"}
}

func (i Installer) Install(ext Extensions) error {
	claude := i.Claude
	if claude == "" {
		claude = "claude"
	}

	for _, p := range ext.Plugins {
		cmd := exec.Command(claude, "plugin", "install", p.Source)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("warn: plugin install %s: %v: %s", p.Name, err, out)
		}
	}

	for _, m := range ext.MCP {
		scope := m.Scope
		if scope == "" {
			scope = "local"
		}
		args := []string{"mcp", "add", m.Name, "--scope", scope, m.Command}
		args = append(args, m.Args...)
		cmd := exec.Command(claude, args...)
		cmd.Env = append(os.Environ(), envPairs(m.Env)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("warn: mcp add %s: %v: %s", m.Name, err, out)
		}
	}
	return nil
}

func envPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for k, v := range env {
		pairs = append(pairs, k+"="+v)
	}
	return pairs
}
```

- [ ] **Schritt 4: Tests ausführen — müssen PASS**

```
go test ./internal/extensions/...
```

Erwartet: `ok  github.com/markus/claude-setup/internal/extensions`

- [ ] **Schritt 5: Commit**

```bash
git add internal/extensions/install.go internal/extensions/install_test.go
git commit -m "feat(extensions): add Installer"
```

---

### Task 3: `deploy.Run()` Integration

**Files:**
- Modify: `internal/deploy/deploy.go`
- Modify: `internal/deploy/deploy_test.go`

- [ ] **Schritt 1: Bestehenden Deploy-Test lesen**

```
cat internal/deploy/deploy_test.go
```

Verstehe den bestehenden Testaufbau (TempDir für source und dest).

- [ ] **Schritt 2: Failing Test schreiben**

Füge am Ende von `deploy_test.go` hinzu:

```go
func TestRunInstallsExtensions(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Fake claude binary
	claudeDir := t.TempDir()
	argsFile := filepath.Join(claudeDir, "calls.txt")
	fakeClaude := filepath.Join(claudeDir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", argsFile)
	os.WriteFile(fakeClaude, []byte(script), 0755)

	// Minimaler base layer
	baseDir := filepath.Join(src, "base")
	os.MkdirAll(baseDir, 0755)
	os.WriteFile(filepath.Join(baseDir, "CLAUDE.md"), []byte("<!-- GENERATED:BEGIN -->\n# Claude\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)
	os.MkdirAll(filepath.Join(baseDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(baseDir, ".claude", "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
	os.WriteFile(filepath.Join(baseDir, "extensions.yaml"), []byte("plugins:\n  - name: superpowers\n    source: claude-plugins-official/superpowers\n"), 0644)

	cfg := config.Config{Profile: "backend"}
	if err := RunWithClaude(src, dst, cfg, fakeClaude); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install claude-plugins-official/superpowers") {
		t.Errorf("plugin not installed, calls: %q", string(data))
	}
}
```

Dazu braucht `deploy_test.go` diese Imports:
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markus/claude-setup/internal/config"
)
```

- [ ] **Schritt 3: Tests ausführen — müssen FAIL**

```
go test ./internal/deploy/...
```

Erwartet: `undefined: RunWithClaude`

- [ ] **Schritt 4: `deploy.go` anpassen**

Ersetze `Run` und füge `RunWithClaude` hinzu. Vollständige neue Version von `internal/deploy/deploy.go`:

```go
package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/markus/claude-setup/internal/compose"
	"github.com/markus/claude-setup/internal/config"
	"github.com/markus/claude-setup/internal/extensions"
)

func Run(sourceDir, destDir string, cfg config.Config) error {
	return RunWithClaude(sourceDir, destDir, cfg, "claude")
}

func RunWithClaude(sourceDir, destDir string, cfg config.Config, claudeBin string) error {
	req := compose.Request{
		SourceDir: sourceDir,
		DestDir:   destDir,
		Profile:   cfg.Profile,
		Flavors:   cfg.Flavors,
	}
	if err := compose.Run(req); err != nil {
		return fmt.Errorf("compose: %w", err)
	}

	if err := copyHooks(sourceDir, destDir); err != nil {
		return fmt.Errorf("hooks: %w", err)
	}

	if err := installExtensions(sourceDir, cfg, claudeBin); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	cfgPath := filepath.Join(destDir, ".claude-setup.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func installExtensions(sourceDir string, cfg config.Config, claudeBin string) error {
	paths := []string{
		filepath.Join(sourceDir, "base", "extensions.yaml"),
		filepath.Join(sourceDir, "profiles", cfg.Profile, "extensions.yaml"),
	}
	for _, flavor := range cfg.Flavors {
		paths = append(paths, filepath.Join(sourceDir, "flavors", flavor, "extensions.yaml"))
	}

	var layers []extensions.Extensions
	for _, path := range paths {
		ext, err := extensions.Load(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		layers = append(layers, ext)
	}

	merged := extensions.Merge(layers)
	return extensions.Installer{Claude: claudeBin}.Install(merged)
}

func copyHooks(src, dst string) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil
	}

	dstHooks := filepath.Join(dst, ".claude", "hooks")
	if err := os.MkdirAll(dstHooks, 0755); err != nil {
		return fmt.Errorf("mkdir hooks: %w", err)
	}

	return filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(hooksDir, path)
		return copyExecutable(path, filepath.Join(dstHooks, rel))
	})
}

func copyExecutable(src, dst string) (err error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
```

- [ ] **Schritt 5: Tests ausführen — müssen PASS**

```
go test ./internal/...
```

Erwartet: alle Pakete `ok`

- [ ] **Schritt 6: Commit**

```bash
git add internal/deploy/deploy.go internal/deploy/deploy_test.go
git commit -m "feat(deploy): integrate extensions install step"
```

---

### Task 4: `base/extensions.yaml` anlegen

**Files:**
- Create: `base/extensions.yaml`

- [ ] **Schritt 1: Datei schreiben**

```yaml
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers
```

- [ ] **Schritt 2: Gesamte Test-Suite ausführen**

```
go test ./...
```

Erwartet: alle Pakete `ok`

- [ ] **Schritt 3: Commit**

```bash
git add base/extensions.yaml
git commit -m "feat(base): declare superpowers plugin in extensions.yaml"
```
