# Agent Permission Mode — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `permission_mode` in `.forgecrate.yaml` konfigurieren, in `settings.json` deployen und via `forgecrate set-permission-mode <mode>` nachträglich ändern.

**Architecture:** Neues Feld `PermissionMode` in `config.Config` wird über `compose.Request` in das zusammengesetzte `settings.json` injiziert. `deploy.PatchPermissionMode` patcht `settings.json` direkt ohne GitHub-Download. Der neue Subcommand `set-permission-mode` kombiniert Patch + Config-Update. `forgecrate init` bekommt `--permission-mode`-Flag mit Default `bypass`.

**Tech Stack:** Go 1.24, cobra, gopkg.in/yaml.v3, encoding/json, crypto/sha256

---

### Task 1: Config — PermissionMode-Feld + Validierung

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Tests schreiben (schlagen fehl)**

Ans Ende von `internal/config/config_test.go` anfügen:

```go
func TestPermissionModeRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	want := config.Config{
		Version:        "1.0",
		Source:         "github.com/jmt-labs/forgecrate",
		Ref:            "main",
		Profile:        "backend",
		Flavors:        []string{"tdd"},
		PermissionMode: "bypass",
	}

	if err := config.Write(path, want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.PermissionMode != want.PermissionMode {
		t.Errorf("PermissionMode: got %q, want %q", got.PermissionMode, want.PermissionMode)
	}
}

func TestPermissionModeOmittedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "p"}
	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "permission_mode") {
		t.Error("permission_mode should be omitted when empty")
	}
}

func TestValidatePermissionMode(t *testing.T) {
	for _, mode := range []string{"bypass", "plan", "ask", "auto"} {
		if err := config.ValidatePermissionMode(mode); err != nil {
			t.Errorf("mode %q should be valid, got %v", mode, err)
		}
	}
	if err := config.ValidatePermissionMode("invalid"); err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := config.Config{PermissionMode: "bypass"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid config: %v", err)
	}

	cfg.PermissionMode = "bad"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid PermissionMode")
	}

	cfg.PermissionMode = ""
	if err := cfg.Validate(); err != nil {
		t.Errorf("empty PermissionMode should be valid: %v", err)
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```
go test ./internal/config/... -run "TestPermissionMode|TestValidate|TestConfigValidate" -v
```

Erwartet: FAIL (`PermissionMode` undefiniert, `ValidatePermissionMode` undefiniert)

- [ ] **Step 3: Implementierung in `internal/config/config.go`**

Import `"fmt"` hinzufügen. Datei ersetzen:

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var ValidPermissionModes = []string{"bypass", "plan", "ask", "auto"}

type Config struct {
	Version        string            `yaml:"version"`
	Source         string            `yaml:"source"`
	Ref            string            `yaml:"ref"`
	Profile        string            `yaml:"profile"`
	Flavors        []string          `yaml:"flavors"`
	PermissionMode string            `yaml:"permission_mode,omitempty"`
	DeployedFiles  map[string]string `yaml:"deployed_files,omitempty"`
}

func ValidatePermissionMode(mode string) error {
	for _, m := range ValidPermissionModes {
		if mode == m {
			return nil
		}
	}
	return fmt.Errorf("ungültiger Modus %q — erlaubt: bypass, plan, ask, auto", mode)
}

func (c Config) Validate() error {
	if c.PermissionMode != "" {
		return ValidatePermissionMode(c.PermissionMode)
	}
	return nil
}

func Read(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		f.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
```

- [ ] **Step 4: Tests ausführen — müssen bestehen**

```
go test ./internal/config/... -v
```

Erwartet: alle PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add PermissionMode field and validation"
```

---

### Task 2: Compose — permissionMode in settings.json injizieren

**Files:**
- Modify: `internal/compose/compose.go`
- Modify: `internal/compose/compose_test.go`
- Modify: `internal/deploy/deploy.go`

- [ ] **Step 1: Tests schreiben (schlagen fehl)**

Ans Ende von `internal/compose/compose_test.go` anfügen. Import `"encoding/json"` hinzufügen:

```go
func TestComposeSettingsInjectsPermissionMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	os.MkdirAll(settingsDir, 0755)
	os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`), 0644)

	req := compose.Request{
		SourceDir:      src,
		DestDir:        dst,
		Profile:        "backend",
		Flavors:        []string{},
		PermissionMode: "bypass",
	}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(content, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
}

func TestComposeSettingsNoPermissionModeWhenEmpty(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	os.MkdirAll(settingsDir, 0755)
	os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`), 0644)

	req := compose.Request{SourceDir: src, DestDir: dst, Profile: "backend", Flavors: []string{}}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}

	if strings.Contains(string(content), "permissionMode") {
		t.Error("permissionMode should not appear when PermissionMode is empty")
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```
go test ./internal/compose/... -run "TestComposeSettingsInjects|TestComposeSettingsNoPerm" -v
```

Erwartet: FAIL (`PermissionMode` in `Request` undefiniert)

- [ ] **Step 3: `compose.Request` erweitern und `ComposeSettings` anpassen**

In `internal/compose/compose.go` — `Request`-Struct um Feld ergänzen:

```go
type Request struct {
	SourceDir      string
	DestDir        string
	Profile        string
	Flavors        []string
	PermissionMode string
	SkipSettings   bool
}
```

Im selben File — den letzten Block von `ComposeSettings` (ab `var v any`) ersetzen:

```go
	var m map[string]any
	if err := json.Unmarshal([]byte(merged), &m); err != nil {
		return nil, fmt.Errorf("merged JSON invalid: %w", err)
	}
	if req.PermissionMode != "" {
		m["permissionMode"] = req.PermissionMode
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return append(out, '\n'), nil
```

- [ ] **Step 4: Tests ausführen — müssen bestehen**

```
go test ./internal/compose/... -v
```

Erwartet: alle PASS

- [ ] **Step 5: `deploy.go` — PermissionMode übergeben + Validierung**

In `internal/deploy/deploy.go` — `RunWithClaude` anpassen:

Nach `func RunWithClaude(...)` als erstes Statement hinzufügen:
```go
	if err := cfg.Validate(); err != nil {
		return err
	}
```

`req`-Initialisierung um `PermissionMode` erweitern:
```go
	req := compose.Request{
		SourceDir:      sourceDir,
		DestDir:        destDir,
		Profile:        cfg.Profile,
		Flavors:        cfg.Flavors,
		PermissionMode: cfg.PermissionMode,
		SkipSettings:   true,
	}
```

- [ ] **Step 6: Alle Tests ausführen**

```
go test ./... 
```

Erwartet: alle PASS

- [ ] **Step 7: Commit**

```bash
git add internal/compose/compose.go internal/compose/compose_test.go internal/deploy/deploy.go
git commit -m "feat(compose): inject permissionMode into settings.json"
```

---

### Task 3: deploy.PatchPermissionMode

**Files:**
- Create: `internal/deploy/patch.go`
- Create: `internal/deploy/patch_test.go`

- [ ] **Step 1: Tests schreiben**

Neue Datei `internal/deploy/patch_test.go` erstellen:

```go
package deploy_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

func TestPatchPermissionMode(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	os.MkdirAll(settingsDir, 0755)

	initial := `{"model":"claude-sonnet-4-6","permissions":{"allow":["Bash"]}}` + "\n"
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(initial), 0644)

	cfg := config.Config{
		DeployedFiles: map[string]string{".claude/settings.json": "sha256:old"},
	}

	if err := deploy.PatchPermissionMode(dir, "bypass", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var m map[string]any
	json.Unmarshal(data, &m)

	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
	if m["model"] != "claude-sonnet-4-6" {
		t.Error("model should be preserved")
	}
	if cfg.DeployedFiles[".claude/settings.json"] == "sha256:old" {
		t.Error("hash should be updated after patch")
	}
}

func TestPatchPermissionModeRemovesKey(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	os.MkdirAll(settingsDir, 0755)

	initial := `{"permissionMode":"bypass","model":"claude-sonnet-4-6"}` + "\n"
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(initial), 0644)

	cfg := config.Config{}
	if err := deploy.PatchPermissionMode(dir, "", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	if strings.Contains(string(data), "permissionMode") {
		t.Error("permissionMode should be removed when mode is empty")
	}
}

func TestPatchPermissionModeMissingSettings(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	err := deploy.PatchPermissionMode(dir, "bypass", &cfg)
	if err == nil {
		t.Error("expected error for missing settings.json")
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```
go test ./internal/deploy/... -run "TestPatch" -v
```

Erwartet: FAIL (`PatchPermissionMode` undefiniert)

- [ ] **Step 3: `internal/deploy/patch.go` erstellen**

```go
package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmt-labs/forgecrate/internal/config"
)

// PatchPermissionMode patcht permissionMode in .claude/settings.json ohne vollständigen Redeploy.
// Aktualisiert außerdem den gespeicherten Hash in cfg.DeployedFiles.
func PatchPermissionMode(destDir string, mode string, cfg *config.Config) error {
	settingsPath := filepath.Join(destDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("settings.json lesen: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("settings.json parsen: %w", err)
	}

	if mode == "" {
		delete(m, "permissionMode")
	} else {
		m["permissionMode"] = mode
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	out = append(out, '\n')

	if err := os.WriteFile(settingsPath, out, 0644); err != nil {
		return fmt.Errorf("settings.json schreiben: %w", err)
	}

	if cfg.DeployedFiles == nil {
		cfg.DeployedFiles = map[string]string{}
	}
	cfg.DeployedFiles[".claude/settings.json"] = hashBytes(out)
	return nil
}
```

- [ ] **Step 4: Tests ausführen — müssen bestehen**

```
go test ./internal/deploy/... -v
```

Erwartet: alle PASS

- [ ] **Step 5: Commit**

```bash
git add internal/deploy/patch.go internal/deploy/patch_test.go
git commit -m "feat(deploy): add PatchPermissionMode for in-place settings update"
```

---

### Task 4: Subcommand `forgecrate set-permission-mode`

**Files:**
- Create: `cmd/forgecrate/set_permission_mode.go`
- Create: `cmd/forgecrate/set_permission_mode_test.go`
- Modify: `cmd/forgecrate/main.go`

- [ ] **Step 1: Tests schreiben**

Neue Datei `cmd/forgecrate/set_permission_mode_test.go` erstellen:

```go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

func TestSetPermissionModeRoundtrip(t *testing.T) {
	dir := t.TempDir()

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
		DeployedFiles: map[string]string{
			".claude/settings.json": "sha256:old",
		},
	}
	cfgPath := filepath.Join(dir, ".forgecrate.yaml")
	config.Write(cfgPath, cfg)

	settingsDir := filepath.Join(dir, ".claude")
	os.MkdirAll(settingsDir, 0755)
	os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`+"\n"), 0644)

	if err := deploy.PatchPermissionMode(dir, "bypass", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}
	cfg.PermissionMode = "bypass"
	config.Write(cfgPath, cfg)

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v", m["permissionMode"])
	}

	got, _ := config.Read(cfgPath)
	if got.PermissionMode != "bypass" {
		t.Errorf("config PermissionMode: got %q", got.PermissionMode)
	}
}

func TestSetPermissionModeValidation(t *testing.T) {
	for _, mode := range []string{"bypass", "plan", "ask", "auto"} {
		if err := config.ValidatePermissionMode(mode); err != nil {
			t.Errorf("mode %q should be valid: %v", mode, err)
		}
	}
	if err := config.ValidatePermissionMode("foo"); err == nil {
		t.Error("expected error for invalid mode")
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen bestehen** (basieren nur auf bereits implementierten Funktionen)

```
go test ./cmd/forgecrate/... -run "TestSetPermissionMode" -v
```

Erwartet: PASS (Tests nutzen `config` und `deploy` direkt)

- [ ] **Step 3: `cmd/forgecrate/set_permission_mode.go` erstellen**

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	"github.com/spf13/cobra"
)

func newSetPermissionModeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-permission-mode <mode>",
		Short: "Setzt den Agent-Berechtigungsmodus (bypass|plan|ask|auto)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := args[0]
			if err := config.ValidatePermissionMode(mode); err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := filepath.Join(cwd, ".forgecrate.yaml")
			cfg, err := config.Read(cfgPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf(".forgecrate.yaml nicht gefunden — erst 'forgecrate init' ausführen")
			} else if err != nil {
				return err
			}

			cfg.PermissionMode = mode

			if err := deploy.PatchPermissionMode(cwd, mode, &cfg); err != nil {
				return err
			}

			if err := config.Write(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("✓ permission_mode: %s\n", mode)
			fmt.Println("✓ .claude/settings.json aktualisiert")
			return nil
		},
	}
}
```

- [ ] **Step 4: Subcommand in `cmd/forgecrate/main.go` registrieren**

Nach `root.AddCommand(newHookCmd())` einfügen:

```go
	root.AddCommand(newSetPermissionModeCmd())
```

- [ ] **Step 5: Bauen**

```
go build ./cmd/forgecrate/...
```

Erwartet: kein Fehler

- [ ] **Step 6: Alle Tests**

```
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 7: Commit**

```bash
git add cmd/forgecrate/set_permission_mode.go cmd/forgecrate/set_permission_mode_test.go cmd/forgecrate/main.go
git commit -m "feat(cmd): add set-permission-mode subcommand"
```

---

### Task 5: `forgecrate init` — `--permission-mode`-Flag + E2E-Tests

**Files:**
- Modify: `cmd/forgecrate/init.go`
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: E2E-Tests schreiben**

Ans Ende von `e2e/e2e_test.go` anfügen. Import `"encoding/json"` hinzufügen:

```go
func TestPermissionModeInSettings(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version:        "1.0",
		Source:         "github.com/jmt-labs/forgecrate",
		Ref:            "main",
		Profile:        "backend",
		Flavors:        []string{},
		PermissionMode: "bypass",
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
}

func TestSetPermissionModeE2E(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	cfgPath := filepath.Join(dst, ".forgecrate.yaml")
	got, err := config.Read(cfgPath)
	if err != nil {
		t.Fatalf("config.Read: %v", err)
	}

	if err := deploy.PatchPermissionMode(dst, "plan", &got); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}
	got.PermissionMode = "plan"
	config.Write(cfgPath, got)

	data, _ := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["permissionMode"] != "plan" {
		t.Errorf("after patch permissionMode: got %v, want plan", m["permissionMode"])
	}

	// Redeploy muss permission_mode aus Config übernehmen
	got2, _ := config.Read(cfgPath)
	if err := deploy.Run(localSource(t), dst, got2); err != nil {
		t.Fatalf("second deploy.Run: %v", err)
	}
	data2, _ := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	var m2 map[string]any
	json.Unmarshal(data2, &m2)
	if m2["permissionMode"] != "plan" {
		t.Errorf("after redeploy permissionMode: got %v, want plan", m2["permissionMode"])
	}
}
```

- [ ] **Step 2: E2E-Tests ausführen — müssen bestehen**

```
go test ./e2e/... -run "TestPermissionMode|TestSetPermissionModeE2E" -v
```

Erwartet: PASS (da `PermissionMode` in Config und Compose bereits implementiert)

- [ ] **Step 3: `--permission-mode`-Flag in `cmd/forgecrate/init.go` hinzufügen**

Variable deklarieren (neben `profile` und `flavors`):

```go
var permissionMode string
```

Flag registrieren (nach den bestehenden `cmd.Flags`-Aufrufen):

```go
cmd.Flags().StringVar(&permissionMode, "permission-mode", "bypass", "Agent-Berechtigungsmodus (bypass|plan|ask|auto)")
```

In `RunE`, nach dem Block `if cmd.Flags().Changed("profile")` einfügen:

```go
			if cmd.Flags().Changed("permission-mode") {
				if err := config.ValidatePermissionMode(permissionMode); err != nil {
					return err
				}
				cfg.PermissionMode = permissionMode
			} else if cfg.PermissionMode == "" {
				cfg.PermissionMode = permissionMode // Default: bypass
			}
```

Import `"github.com/jmt-labs/forgecrate/internal/config"` ist bereits vorhanden.

- [ ] **Step 4: Bauen + alle Tests**

```
go build ./cmd/forgecrate/... && go test ./...
```

Erwartet: kein Fehler, alle PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/forgecrate/init.go e2e/e2e_test.go
git commit -m "feat(init): add --permission-mode flag, default bypass"
```

---

### Abschluss-Verifikation

- [ ] **Vollständiger Test-Lauf**

```
go test ./... -race
```

Erwartet: alle PASS, keine Race Conditions

- [ ] **Binary bauen und manuell testen**

```
go build -o /tmp/forgecrate ./cmd/forgecrate && /tmp/forgecrate set-permission-mode --help
```

Erwartet: Hilfetext mit `<mode>` und `bypass|plan|ask|auto`

- [ ] **Final Commit / PR erstellen**
