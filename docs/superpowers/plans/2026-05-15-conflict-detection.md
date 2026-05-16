# Konflikterkennung mit Hash-Tracking — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `claude-setup update` erkennt, wenn der Nutzer eine verwaltete Datei (Settings, Hooks) seit dem letzten Deploy manuell geändert hat, und fragt vor dem Überschreiben nach.

**Architecture:** SHA-256-Hashes in `deployed_files`-Feld von `.claude-setup.yaml`. Drei-Wege-Vergleich (disk vs. stored vs. new) in neuer `deployFile()`-Funktion. `compose.ComposeSettings()` (neuer Export) liefert den neuen Settings-Inhalt bevor er geschrieben wird. `copyHooks` nutzt `deployFile` statt direktem `copyFile`. Kein Umbau der öffentlichen Deploy-Signatur nötig.

**Tech Stack:** Go (`crypto/sha256`, `io`, `bufio`), bestehende `deploy.Run()` / `compose.Run()` E2E-Infrastruktur, `testing`.

---

### Task 1: Config.DeployedFiles Feld

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Fehlschlagenden Test schreiben**

In `internal/config/config_test.go` (neu anlegen falls nicht vorhanden):

```go
package config_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/jmt-labs/claude-setup/internal/config"
)

func TestDeployedFilesRoundtrip(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, ".claude-setup.yaml")

    cfg := config.Config{
        Version: "1.0",
        Source:  "github.com/jmt-labs/claude-setup",
        Ref:     "main",
        Profile: "backend",
        Flavors: []string{"tdd"},
        DeployedFiles: map[string]string{
            ".claude/settings.json":        "sha256:abc123",
            ".claude/hooks/pre-tool.sh":    "sha256:def456",
        },
    }

    if err := config.Write(path, cfg); err != nil {
        t.Fatalf("Write: %v", err)
    }

    got, err := config.Read(path)
    if err != nil {
        t.Fatalf("Read: %v", err)
    }

    if got.DeployedFiles[".claude/settings.json"] != "sha256:abc123" {
        t.Errorf("settings.json hash lost: %q", got.DeployedFiles[".claude/settings.json"])
    }
    if got.DeployedFiles[".claude/hooks/pre-tool.sh"] != "sha256:def456" {
        t.Errorf("pre-tool.sh hash lost: %q", got.DeployedFiles[".claude/hooks/pre-tool.sh"])
    }
}

func TestDeployedFilesOmittedWhenEmpty(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, ".claude-setup.yaml")

    cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "p"}
    if err := config.Write(path, cfg); err != nil {
        t.Fatalf("Write: %v", err)
    }

    data, _ := os.ReadFile(path)
    if strings.Contains(string(data), "deployed_files") {
        t.Error("deployed_files should be omitted when empty")
    }
}
```

Fehlende Imports ergänzen (`strings`).

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/config/... -v
```

Erwartet: FAIL (Feld existiert nicht)

- [ ] **Step 3: DeployedFiles-Feld hinzufügen**

In `internal/config/config.go` das Struct erweitern:

```go
type Config struct {
    Version       string            `yaml:"version"`
    Source        string            `yaml:"source"`
    Ref           string            `yaml:"ref"`
    Profile       string            `yaml:"profile"`
    Flavors       []string          `yaml:"flavors"`
    DeployedFiles map[string]string `yaml:"deployed_files,omitempty"`
}
```

- [ ] **Step 4: Test ausführen — muss bestehen**

```bash
go test ./internal/config/... -v
```

Erwartet: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: Config.DeployedFiles für Hash-Tracking"
```

---

### Task 2: Hash-Utility

**Files:**
- Create: `internal/deploy/hash.go`
- Create: `internal/deploy/hash_test.go`

- [ ] **Step 1: Fehlschlagenden Test schreiben**

Inhalt von `internal/deploy/hash_test.go`:

```go
package deploy

import (
    "os"
    "path/filepath"
    "testing"
)

func TestHashBytes(t *testing.T) {
    h := hashBytes([]byte("hello"))
    if h != "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
        t.Errorf("unexpected hash: %s", h)
    }
}

func TestHashBytesEmpty(t *testing.T) {
    h := hashBytes([]byte{})
    if h != "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
        t.Errorf("unexpected hash: %s", h)
    }
}

func TestHashFile(t *testing.T) {
    dir := t.TempDir()
    f := filepath.Join(dir, "test.txt")
    os.WriteFile(f, []byte("hello"), 0644)

    h, err := hashFile(f)
    if err != nil {
        t.Fatalf("hashFile: %v", err)
    }
    if h != hashBytes([]byte("hello")) {
        t.Errorf("hashFile != hashBytes for same content")
    }
}

func TestHashFileMissing(t *testing.T) {
    _, err := hashFile("/no/such/file")
    if err == nil {
        t.Error("expected error for missing file")
    }
}
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/deploy/... -run TestHash -v
```

Erwartet: FAIL (Funktionen nicht definiert)

- [ ] **Step 3: hash.go implementieren**

Inhalt von `internal/deploy/hash.go`:

```go
package deploy

import (
    "crypto/sha256"
    "fmt"
    "os"
)

func hashBytes(data []byte) string {
    sum := sha256.Sum256(data)
    return fmt.Sprintf("sha256:%x", sum)
}

func hashFile(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return hashBytes(data), nil
}
```

- [ ] **Step 4: Test ausführen — muss bestehen**

```bash
go test ./internal/deploy/... -run TestHash -v
```

Erwartet: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/deploy/hash.go internal/deploy/hash_test.go
git commit -m "feat: SHA-256 Hash-Utilities für Konflikterkennung"
```

---

### Task 3: deployFile — Drei-Wege-Vergleich

**Files:**
- Create: `internal/deploy/deployfile.go`
- Create: `internal/deploy/deployfile_test.go`

- [ ] **Step 1: Fehlschlagenden Test schreiben**

Inhalt von `internal/deploy/deployfile_test.go`:

```go
package deploy

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/jmt-labs/claude-setup/internal/config"
)

func writeFile(t *testing.T, path, content string) {
    t.Helper()
    os.MkdirAll(filepath.Dir(path), 0755)
    os.WriteFile(path, []byte(content), 0644)
}

// Fall 1: disk == stored, new == disk → nichts tun
func TestDeployFileNoChange(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "original")

    cfg := &config.Config{
        DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
    }

    err := deployFile(dst, "file.txt", []byte("original"), cfg, &strings.Builder{}, strings.NewReader(""))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "original" {
        t.Errorf("file should be unchanged: %q", data)
    }
}

// Fall 2: disk == stored, new != disk → einfach überschreiben (kein Prompt)
func TestDeployFileCleanUpdate(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "original")

    cfg := &config.Config{
        DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
    }

    out := &strings.Builder{}
    err := deployFile(dst, "file.txt", []byte("updated"), cfg, out, strings.NewReader(""))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "updated" {
        t.Errorf("expected updated content, got %q", data)
    }
    if strings.Contains(out.String(), "KONFLIKT") {
        t.Error("clean update should not show conflict")
    }
    if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("updated")) {
        t.Error("hash not updated in cfg")
    }
}

// Fall 3: disk != stored, new == disk → Nutzer hat geändert, neue Version identisch → nichts tun
func TestDeployFileUserChangedSameAsNew(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "user-modified")

    cfg := &config.Config{
        DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
    }

    err := deployFile(dst, "file.txt", []byte("user-modified"), cfg, &strings.Builder{}, strings.NewReader(""))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "user-modified" {
        t.Errorf("file should be preserved: %q", data)
    }
}

// Fall 4a: Konflikt → Nutzer wählt behalten
func TestDeployFileConflictKeep(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "user-modified")

    cfg := &config.Config{
        DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
    }

    out := &strings.Builder{}
    err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("b\n"))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "user-modified" {
        t.Errorf("file should be kept: %q", data)
    }
    if !strings.Contains(out.String(), "KONFLIKT") {
        t.Error("conflict should be reported")
    }
}

// Fall 4b: Konflikt → Nutzer wählt überschreiben
func TestDeployFileConflictOverwrite(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "user-modified")

    cfg := &config.Config{
        DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
    }

    out := &strings.Builder{}
    err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("ü\n"))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "remote-new" {
        t.Errorf("file should be overwritten: %q", data)
    }
    if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("remote-new")) {
        t.Error("hash not updated after overwrite")
    }
}

// Migration: kein stored hash → einfach überschreiben
func TestDeployFileMissingStoredHash(t *testing.T) {
    dir := t.TempDir()
    dst := filepath.Join(dir, "file.txt")
    writeFile(t, dst, "user-modified")

    cfg := &config.Config{DeployedFiles: map[string]string{}}

    out := &strings.Builder{}
    err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader(""))
    if err != nil {
        t.Fatalf("deployFile: %v", err)
    }
    data, _ := os.ReadFile(dst)
    if string(data) != "remote-new" {
        t.Errorf("migration: expected overwrite, got %q", data)
    }
    if strings.Contains(out.String(), "KONFLIKT") {
        t.Error("migration should not show conflict")
    }
}
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/deploy/... -run TestDeployFile -v
```

Erwartet: FAIL (deployFile nicht definiert)

- [ ] **Step 3: deployfile.go implementieren**

Inhalt von `internal/deploy/deployfile.go`:

```go
package deploy

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/jmt-labs/claude-setup/internal/config"
)

func deployFile(dstPath, rel string, newContent []byte, cfg *config.Config, w io.Writer, r io.Reader) error {
    if cfg.DeployedFiles == nil {
        cfg.DeployedFiles = map[string]string{}
    }

    hashNew := hashBytes(newContent)

    hashDisk := ""
    if diskData, err := os.ReadFile(dstPath); err == nil {
        hashDisk = hashBytes(diskData)
    }

    hashStored, hasStored := cfg.DeployedFiles[rel]

    // Fall: Datei existiert nicht → einfach schreiben
    if hashDisk == "" {
        return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
    }

    // Migration: kein stored hash → einfach überschreiben
    if !hasStored {
        return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
    }

    diskUnchanged := hashDisk == hashStored
    newSameAsDisk := hashNew == hashDisk

    if diskUnchanged {
        if newSameAsDisk {
            // Fall 1: unverändert, neue Version identisch → nichts tun
            return nil
        }
        // Fall 2: unverändert, neue Version verschieden → überschreiben
        return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
    }

    // Nutzer hat geändert
    if newSameAsDisk {
        // Fall 3: Nutzer hat gleich geändert wie neue Version → nichts tun
        return nil
    }

    // Fall 4: echter Konflikt
    diskData, _ := os.ReadFile(dstPath)
    fmt.Fprintf(w, "\nKONFLIKT: %s\n", rel)
    fmt.Fprintf(w, "  Deine Version: %s\n", firstLine(diskData))
    fmt.Fprintf(w, "  Neue Version:  %s\n", firstLine(newContent))
    fmt.Fprintf(w, "  [ü]berschreiben / [b]ehalten (Standard: behalten): ")

    scanner := bufio.NewScanner(r)
    scanner.Scan()
    answer := strings.TrimSpace(scanner.Text())

    if answer == "ü" || answer == "u" {
        return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
    }
    // behalten: Hash der Nutzer-Version speichern
    cfg.DeployedFiles[rel] = hashDisk
    return nil
}

func writeAndRecord(dstPath, rel string, content []byte, hash string, cfg *config.Config) error {
    if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
        return err
    }
    if err := os.WriteFile(dstPath, content, 0644); err != nil {
        return err
    }
    cfg.DeployedFiles[rel] = hash
    return nil
}

func firstLine(data []byte) string {
    s := strings.TrimSpace(string(data))
    if idx := strings.IndexByte(s, '\n'); idx >= 0 {
        s = s[:idx]
    }
    if len(s) > 80 {
        s = s[:80] + "…"
    }
    return s
}
```

Fehlenden Import `path/filepath` ergänzen.

- [ ] **Step 4: Test ausführen — muss bestehen**

```bash
go test ./internal/deploy/... -run TestDeployFile -v
```

Erwartet: alle 6 Tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/deploy/deployfile.go internal/deploy/deployfile_test.go
git commit -m "feat: deployFile — Drei-Wege-Konflikterkennung mit Hash-Tracking"
```

---

### Task 4: compose.ComposeSettings exportieren

**Files:**
- Modify: `internal/compose/compose.go`
- Modify: `internal/compose/compose_test.go` (falls vorhanden, sonst neu)

- [ ] **Step 1: Fehlschlagenden Test schreiben**

In `internal/compose/compose_test.go` (oder ergänzen):

```go
func TestComposeSettingsReturnsContent(t *testing.T) {
    dir := t.TempDir()
    // minimales base settings.json anlegen
    settingsDir := filepath.Join(dir, "base", ".claude")
    os.MkdirAll(settingsDir, 0755)
    os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)

    req := Request{SourceDir: dir, DestDir: t.TempDir(), Profile: "backend", Flavors: []string{}}
    content, err := ComposeSettings(req)
    if err != nil {
        t.Fatalf("ComposeSettings: %v", err)
    }
    if !strings.Contains(string(content), "claude-sonnet-4-6") {
        t.Errorf("expected model in content, got: %s", content)
    }
}
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/compose/... -run TestComposeSettings -v
```

Erwartet: FAIL (`ComposeSettings` nicht definiert)

- [ ] **Step 3: ComposeSettings exportieren + SkipSettings in Request**

`compose.Request` erweitern:

```go
type Request struct {
    SourceDir    string
    DestDir      string
    Profile      string
    Flavors      []string
    SkipSettings bool
}
```

`composeJSON` in zwei Teile teilen:

```go
// ComposeSettings berechnet den gemergten settings.json-Inhalt ohne ihn zu schreiben.
func ComposeSettings(req Request) ([]byte, error) {
    basePath := filepath.Join(req.SourceDir, "base", ".claude", "settings.json")
    data, err := os.ReadFile(basePath)
    if err != nil {
        return nil, err
    }
    merged := string(data)

    profilePath := filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "settings.json")
    if override, err := os.ReadFile(profilePath); err == nil {
        merged, err = DeepMergeJSON(merged, string(override))
        if err != nil {
            return nil, err
        }
    }

    overridePath := filepath.Join(req.DestDir, ".claude", "overrides", "settings.override.json")
    if override, err := os.ReadFile(overridePath); err == nil {
        merged, err = DeepMergeJSON(merged, string(override))
        if err != nil {
            return nil, err
        }
    }

    var v any
    if err := json.Unmarshal([]byte(merged), &v); err != nil {
        return nil, fmt.Errorf("merged JSON invalid: %w", err)
    }
    out, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("marshal: %w", err)
    }
    return append(out, '\n'), nil
}

func composeJSON(req Request) error {
    if req.SkipSettings {
        return nil
    }
    content, err := ComposeSettings(req)
    if err != nil {
        return err
    }
    dst := filepath.Join(req.DestDir, ".claude", "settings.json")
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return fmt.Errorf("mkdir: %w", err)
    }
    return os.WriteFile(dst, content, 0644)
}
```

Die alte `composeJSON`-Implementierung wird durch diesen Aufruf ersetzt.

- [ ] **Step 4: Test ausführen — muss bestehen**

```bash
go test ./internal/compose/... -v
```

Erwartet: alle PASS (inkl. bestehende Tests)

- [ ] **Step 5: Commit**

```bash
git add internal/compose/compose.go internal/compose/compose_test.go
git commit -m "feat: compose.ComposeSettings exportiert, SkipSettings-Flag"
```

---

### Task 5: deploy.go Integration

**Files:**
- Modify: `internal/deploy/deploy.go`

- [ ] **Step 1: Fehlschlagenden Test schreiben**

In `internal/deploy/deploy_test.go` (neu anlegen):

```go
package deploy_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/jmt-labs/claude-setup/internal/config"
    "github.com/jmt-labs/claude-setup/internal/deploy"
)

func TestDeployTracksSettingsHash(t *testing.T) {
    src := t.TempDir()
    dst := t.TempDir()
    setupMinimalSource(t, src)

    cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
    if err := deploy.Run(src, dst, cfg); err != nil {
        t.Fatalf("first deploy: %v", err)
    }

    written, err := config.Read(filepath.Join(dst, ".claude-setup.yaml"))
    if err != nil {
        t.Fatalf("read config: %v", err)
    }
    if written.DeployedFiles[".claude/settings.json"] == "" {
        t.Error("settings.json hash not tracked after deploy")
    }
}

func setupMinimalSource(t *testing.T, src string) {
    t.Helper()
    // base settings.json
    settingsDir := filepath.Join(src, "base", ".claude")
    os.MkdirAll(settingsDir, 0755)
    os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
    // base CLAUDE.md
    os.WriteFile(filepath.Join(src, "base", "CLAUDE.md"), []byte("# Base\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)
}
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/deploy/... -run TestDeployTracksSettingsHash -v
```

Erwartet: FAIL (Hash nicht befüllt)

- [ ] **Step 3: deploy.go anpassen**

`RunWithClaude` so umbauen, dass:
1. `compose.ComposeSettings(req)` aufgerufen wird, um `hash_new` zu berechnen
2. `compose.Run(req mit SkipSettings: true)` aufgerufen wird
3. `deployFile(settingsPath, ".claude/settings.json", settingsContent, &cfg, os.Stdout, os.Stdin)` schreibt settings mit Konflikt-Check
4. `copyHooks` intern `deployFile` nutzt (Signatur erweitern: `copyHooks(src, dst string, cfg *config.Config) error`)
5. `cfg` als Pointer weitergegeben wird damit DeployedFiles aktualisiert werden
6. Am Ende `config.Write(cfgPath, cfg)` schreibt die aktualisierten Hashes

```go
func RunWithClaude(sourceDir, destDir string, cfg config.Config, claudeBin string) error {
    req := compose.Request{
        SourceDir:    sourceDir,
        DestDir:      destDir,
        Profile:      cfg.Profile,
        Flavors:      cfg.Flavors,
        SkipSettings: true,
    }

    // Settings: Inhalt berechnen, dann konflikt-sicher schreiben
    settingsContent, err := compose.ComposeSettings(req)
    if err != nil {
        return fmt.Errorf("compose settings: %w", err)
    }
    settingsPath := filepath.Join(destDir, ".claude", "settings.json")
    if err := deployFile(settingsPath, ".claude/settings.json", settingsContent, &cfg, os.Stdout, os.Stdin); err != nil {
        return fmt.Errorf("settings: %w", err)
    }

    if err := compose.Run(req); err != nil {
        return fmt.Errorf("compose: %w", err)
    }

    if err := copyHooks(sourceDir, destDir, &cfg); err != nil {
        return fmt.Errorf("hooks: %w", err)
    }

    if err := installExtensions(sourceDir, destDir, cfg, claudeBin); err != nil {
        return fmt.Errorf("extensions: %w", err)
    }

    if err := copySkills(sourceDir, destDir, cfg); err != nil {
        return fmt.Errorf("skills: %w", err)
    }

    cfgPath := filepath.Join(destDir, ".claude-setup.yaml")
    if err := config.Write(cfgPath, cfg); err != nil {
        return fmt.Errorf("write config: %w", err)
    }

    return nil
}
```

`copyHooks` anpassen — Signatur und Interna:

```go
func copyHooks(src, dst string, cfg *config.Config) error {
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
        dstPath := filepath.Join(dstHooks, rel)

        // Inhalt lesen für deployFile
        content, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("read hook %s: %w", rel, err)
        }

        // Executable-Permissions nach dem Schreiben setzen
        relKey := filepath.Join(".claude", "hooks", rel)
        if err := deployFile(dstPath, relKey, content, cfg, os.Stdout, os.Stdin); err != nil {
            return err
        }
        return os.Chmod(dstPath, 0755)
    })
}
```

- [ ] **Step 4: Alle Tests ausführen — müssen bestehen**

```bash
go test ./... -v 2>&1 | tail -30
```

Erwartet: alle PASS

- [ ] **Step 5: Commit**

```bash
git add internal/deploy/
git commit -m "feat: deploy.go nutzt deployFile für settings + hooks (Hash-Tracking)"
```

---

### Task 6: E2E-Test für Konfliktszenario

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Fehlschlagenden E2E-Test schreiben**

Am Ende von `e2e/e2e_test.go` einfügen:

```go
func TestConflictDetectionTracksHashes(t *testing.T) {
    dst := t.TempDir()
    cfg := config.Config{
        Version: "1.0",
        Source:  "github.com/jmt-labs/claude-setup",
        Ref:     "main",
        Profile: "backend",
        Flavors: []string{},
    }

    // Erster Deploy — befüllt deployed_files
    if err := deploy.Run(localSource(t), dst, cfg); err != nil {
        t.Fatalf("first deploy: %v", err)
    }

    yamlPath := filepath.Join(dst, ".claude-setup.yaml")
    written, err := config.Read(yamlPath)
    if err != nil {
        t.Fatalf("read config: %v", err)
    }
    if written.DeployedFiles[".claude/settings.json"] == "" {
        t.Error("settings.json hash not tracked after first deploy")
    }
    if written.DeployedFiles[".claude/hooks/prompt-submit.sh"] == "" {
        t.Error("prompt-submit.sh hash not tracked after first deploy")
    }

    // Zweiter Deploy ohne Änderung — Hashes bleiben stabil
    if err := deploy.Run(localSource(t), dst, written); err != nil {
        t.Fatalf("second deploy: %v", err)
    }
    after, _ := config.Read(yamlPath)
    if after.DeployedFiles[".claude/settings.json"] != written.DeployedFiles[".claude/settings.json"] {
        t.Error("hash changed on clean second deploy")
    }
}
```

- [ ] **Step 2: Test ausführen — muss bestehen**

```bash
go test ./e2e/... -run TestConflictDetectionTracksHashes -v
```

Erwartet: PASS (die Logik ist bereits in Task 5 implementiert)

- [ ] **Step 3: Alle Tests ausführen**

```bash
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 4: Commit**

```bash
git add e2e/e2e_test.go
git commit -m "test: E2E-Test für Hash-Tracking nach Deploy"
```

---

### Task 7: Abschluss

- [ ] **Step 1: Alle Tests ein letztes Mal ausführen**

```bash
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 2: Plan-Pfad in Issue #11 ergänzen**

Kommentar in Issue #11: "Plan fertig: `docs/superpowers/plans/2026-05-15-conflict-detection.md`"

- [ ] **Step 3: PR erstellen**

```bash
gh pr create \
  --title "feat: Konflikterkennung mit Hash-Tracking (#11)" \
  --body "$(cat <<'EOF'
## Was

Implementiert Drei-Wege-Konflikterkennung für verwaltete Dateien (settings.json, Hooks).

## Warum

Nutzer-Änderungen an verwalteten Dateien wurden bei `claude-setup update` stillschweigend überschrieben.

## Wie

- `Config.DeployedFiles` (SHA-256-Hashes in `.claude-setup.yaml`)
- `deployFile()`: Drei-Wege-Vergleich hash_disk / hash_stored / hash_new mit interaktivem Prompt bei Konflikt
- `compose.ComposeSettings()`: exportiert Settings-Berechnung ohne Disk-Write
- `compose.Request.SkipSettings`: deploy.go schreibt settings.json selbst (konflikt-sicher)
- Hooks nutzen ebenfalls `deployFile`

## Wie getestet

- Unit-Tests: alle 4 Konfliktfälle + Migration
- E2E: Hash-Tracking nach erstem und zweitem Deploy
- Alle bestehenden Tests unverändert grün

Closes #11
EOF
)"
```
