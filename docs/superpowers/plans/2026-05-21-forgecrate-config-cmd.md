# forgecrate config — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Neuen `forgecrate config`-Command implementieren, der Profile und Flavors per interaktivem Pfeil-Wizard konfiguriert und danach sofort deployed.

**Architecture:** Cobra-Command (`newConfigCmd`) lädt Source-Repo, ruft testbare Kernfunktion `configInteractive` auf (listet Optionen, ruft `promptFn` auf, schreibt YAML), dann `deploy.Run`. `promptFn` ist injizierbar — in Tests ein Stub, in Production `huhPrompt` mit `charmbracelet/huh`. `listDirs` ist bereits in `list.go` definiert und im selben Package nutzbar.

**Tech Stack:** Go 1.24, cobra, charmbracelet/huh (Select + MultiSelect), bestehende interne Packages `config`, `deploy`, `github`.

---

## Dateien

| Aktion | Datei |
|---|---|
| Erstellen | `cmd/forgecrate/config.go` |
| Erstellen | `cmd/forgecrate/config_test.go` |
| Ändern | `cmd/forgecrate/main.go` — eine Zeile: `root.AddCommand(newConfigCmd())` |
| Ändern | `go.mod` / `go.sum` — via `go get` |

---

### Task 1: `charmbracelet/huh` Dependency hinzufügen

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Dependency holen**

```bash
go get github.com/charmbracelet/huh
```

- [ ] **Step 2: Prüfen ob go.mod aktualisiert wurde**

```bash
grep charmbracelet go.mod
```

Erwartete Ausgabe: eine Zeile mit `github.com/charmbracelet/huh v...`

- [ ] **Step 3: Build-Check (keine neuen Fehler)**

```bash
go build ./...
```

Erwartete Ausgabe: kein Output (erfolgreich).

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore(deps): add charmbracelet/huh for interactive prompts"
```

---

### Task 2: Fehlschlagende Tests schreiben

**Files:**
- Create: `cmd/forgecrate/config_test.go`

- [ ] **Step 1: Testdatei anlegen**

Datei `cmd/forgecrate/config_test.go` mit folgendem Inhalt erstellen:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func TestConfigInteractive_WritesUpdatedConfig(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, ".forgecrate.yaml"), []byte(
		"version: \"1.0\"\nsource: github.com/jmt-labs/forgecrate\nref: main\nprofile: backend\nflavors:\n  - tdd\n",
	), 0644); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"backend", "frontend", "fullstack"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "profiles", p), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tdd", "strict-review", "github"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "flavors", f), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
	}

	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		for _, p := range []string{"backend", "frontend", "fullstack"} {
			found := false
			for _, got := range profiles {
				if got == p {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("profiles missing %q, got %v", p, profiles)
			}
		}
		return "frontend", []string{"strict-review"}, nil
	}

	got, err := configInteractive(dir, srcDir, cfg, stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Profile != "frontend" {
		t.Errorf("returned profile = %q, want frontend", got.Profile)
	}
	if len(got.Flavors) != 1 || got.Flavors[0] != "strict-review" {
		t.Errorf("returned flavors = %v, want [strict-review]", got.Flavors)
	}

	written, err := config.Read(filepath.Join(dir, ".forgecrate.yaml"))
	if err != nil {
		t.Fatalf("read back config: %v", err)
	}
	if written.Profile != "frontend" {
		t.Errorf("written profile = %q, want frontend", written.Profile)
	}
	if len(written.Flavors) != 1 || written.Flavors[0] != "strict-review" {
		t.Errorf("written flavors = %v, want [strict-review]", written.Flavors)
	}
}

func TestConfigInteractive_EmptyProfiles(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	// Kein profiles/-Verzeichnis im srcDir

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		t.Fatal("prompt should not be called when profiles are empty")
		return "", nil, nil
	}

	_, err := configInteractive(dir, srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error for empty profiles, got nil")
	}
}

func TestConfigInteractive_PromptError(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()

	for _, p := range []string{"backend"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "profiles", p), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tdd"} {
		if err := os.MkdirAll(filepath.Join(srcDir, "flavors", f), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{Profile: "backend"}
	stub := func(profiles, flavors []string, cur config.Config) (string, []string, error) {
		return "", nil, fmt.Errorf("abgebrochen")
	}

	_, err := configInteractive(dir, srcDir, cfg, stub)
	if err == nil {
		t.Fatal("expected error from prompt, got nil")
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen FEHLSCHLAGEN**

```bash
go test ./cmd/forgecrate/... -run TestConfigInteractive -v
```

Erwartete Ausgabe: `FAIL` mit `undefined: configInteractive`

- [ ] **Step 3: Commit**

```bash
git add cmd/forgecrate/config_test.go
git commit -m "test(config): add failing tests for configInteractive"
```

---

### Task 3: `configInteractive` implementieren

**Files:**
- Create: `cmd/forgecrate/config.go`

- [ ] **Step 1: `config.go` erstellen**

Datei `cmd/forgecrate/config.go` mit folgendem Inhalt erstellen:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

// promptFn wird in Tests durch einen Stub ersetzt, in Production durch huhPrompt.
type promptFn func(profiles, flavors []string, cur config.Config) (string, []string, error)

// configInteractive liest verfügbare Optionen aus srcDir, ruft prompt auf und
// schreibt die aktualisierte Konfiguration nach cwd/.forgecrate.yaml.
func configInteractive(cwd, srcDir string, cfg config.Config, prompt promptFn) (config.Config, error) {
	profiles, err := listDirs(filepath.Join(srcDir, "profiles"))
	if err != nil {
		return config.Config{}, fmt.Errorf("profile-Liste lesen: %w", err)
	}
	if len(profiles) == 0 {
		return config.Config{}, fmt.Errorf("keine Profile im Source-Repo gefunden")
	}

	flavors, err := listDirs(filepath.Join(srcDir, "flavors"))
	if err != nil {
		return config.Config{}, fmt.Errorf("flavor-Liste lesen: %w", err)
	}
	if len(flavors) == 0 {
		return config.Config{}, fmt.Errorf("keine Flavors im Source-Repo gefunden")
	}

	newProfile, newFlavors, err := prompt(profiles, flavors, cfg)
	if err != nil {
		return config.Config{}, err
	}

	cfg.Profile = newProfile
	cfg.Flavors = newFlavors
	if err := config.Write(filepath.Join(cwd, ".forgecrate.yaml"), cfg); err != nil {
		return config.Config{}, fmt.Errorf("config schreiben: %w", err)
	}
	return cfg, nil
}

// huhPrompt zeigt einen interaktiven Pfeil-Wizard für Profil und Flavors.
func huhPrompt(profiles, flavors []string, cur config.Config) (string, []string, error) {
	newProfile := cur.Profile
	newFlavors := make([]string, len(cur.Flavors))
	copy(newFlavors, cur.Flavors)

	profileOpts := make([]huh.Option[string], len(profiles))
	for i, p := range profiles {
		profileOpts[i] = huh.NewOption(p, p)
	}

	flavorOpts := make([]huh.Option[string], len(flavors))
	for i, f := range flavors {
		flavorOpts[i] = huh.NewOption(f, f)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Profil").
				Options(profileOpts...).
				Value(&newProfile),
			huh.NewMultiSelect[string]().
				Title("Flavors  (Leertaste = toggle)").
				Options(flavorOpts...).
				Value(&newFlavors),
		),
	)

	if err := form.Run(); err != nil {
		return "", nil, err
	}
	return newProfile, newFlavors, nil
}

// newConfigCmd gibt den cobra-Command zurück. Er lädt das Source-Repo,
// ruft configInteractive auf und deployed anschließend.
func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Profil und Flavors interaktiv konfigurieren",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := filepath.Join(cwd, ".forgecrate.yaml")
			cfg, err := config.Read(cfgPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("kein forgecrate-Repo. Zuerst `forgecrate init` ausführen")
			}
			if err != nil {
				return err
			}

			fmt.Printf("Fetching jmt-labs/forgecrate@%s ...\n", cfg.Ref)
			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(srcDir) }()

			client := gh.Default()
			if err := client.Download("jmt-labs", "forgecrate", cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			updatedCfg, err := configInteractive(cwd, srcDir, cfg, huhPrompt)
			if err != nil {
				return err
			}

			fmt.Printf("Deploying profile=%s flavors=%v ...\n", updatedCfg.Profile, updatedCfg.Flavors)
			if err := deploy.Run(srcDir, cwd, updatedCfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen BESTEHEN**

```bash
go test ./cmd/forgecrate/... -run TestConfigInteractive -v
```

Erwartete Ausgabe:
```
--- PASS: TestConfigInteractive_WritesUpdatedConfig
--- PASS: TestConfigInteractive_EmptyProfiles
--- PASS: TestConfigInteractive_PromptError
PASS
```

- [ ] **Step 3: Build-Check**

```bash
go build ./...
```

Erwartete Ausgabe: kein Output.

- [ ] **Step 4: Commit**

```bash
git add cmd/forgecrate/config.go
git commit -m "feat(config): implement configInteractive and huhPrompt"
```

---

### Task 4: Command in `main.go` registrieren

**Files:**
- Modify: `cmd/forgecrate/main.go`

- [ ] **Step 1: `newConfigCmd()` in main.go eintragen**

In `cmd/forgecrate/main.go` die Zeile `root.AddCommand(newDescribeCmd())` suchen und **danach** einfügen:

```go
root.AddCommand(newConfigCmd())
```

Der Block sieht danach so aus:

```go
root.AddCommand(newInitCmd())
root.AddCommand(newUpdateCmd())
root.AddCommand(newListCmd())
root.AddCommand(newDescribeCmd())
root.AddCommand(newConfigCmd())
root.AddCommand(newHookCmd())
```

- [ ] **Step 2: Gesamtes Test-Suite ausführen**

```bash
go test ./...
```

Erwartete Ausgabe: `ok` für alle Packages, keine Failures.

- [ ] **Step 3: Build-Check mit Version**

```bash
go build -o /tmp/forgecrate-test ./cmd/forgecrate && /tmp/forgecrate-test --help
```

Erwartete Ausgabe enthält: `config      Profil und Flavors interaktiv konfigurieren`

- [ ] **Step 4: Commit**

```bash
git add cmd/forgecrate/main.go
git commit -m "feat(config): register forgecrate config command"
```

---

### Task 5: Abschluss-Verification

- [ ] **Step 1: Alle Tests grün**

```bash
go test ./... -count=1
```

Erwartete Ausgabe: alle Packages `ok`.

- [ ] **Step 2: Build für alle Zielplattformen**

```bash
GOOS=linux GOARCH=amd64 go build ./cmd/forgecrate && \
GOOS=darwin GOARCH=arm64 go build ./cmd/forgecrate
```

Erwartete Ausgabe: kein Output (erfolgreich).

- [ ] **Step 3: `forgecrate config --help` prüfen**

```bash
go run ./cmd/forgecrate config --help
```

Erwartete Ausgabe:
```
Profil und Flavors interaktiv konfigurieren

Usage:
  forgecrate config [flags]
```

- [ ] **Step 4: Lint**

```bash
golangci-lint run ./...
```

Erwartete Ausgabe: keine Fehler.

- [ ] **Step 5: Finaler Commit falls nötig**

Falls noch unstaged Changes vorhanden:
```bash
git add -p
git commit -m "fix(config): address lint findings"
```
