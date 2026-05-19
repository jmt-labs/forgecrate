# forgecrate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ein globales Go-Binary (`forgecrate`) das ein Layer-basiertes Claude-Konfigurations-Setup (CLAUDE.md, AGENTS.md, settings.json, Skills, Hooks) von GitHub in beliebige Ziel-Repos deployt.

**Architecture:** Layer-System mit drei Ebenen (base → profile → flavors), compositioniert durch das Binary und in das Ziel-Repo geschrieben. Lokale Overrides (`<!-- CUSTOM:BEGIN/END -->` in Markdown, `overrides/` für Skills) werden nie überschrieben. Hooks in settings.json erzwingen Pflicht-Skills zur Laufzeit.

**Tech Stack:** Go 1.22 · github.com/spf13/cobra · gopkg.in/yaml.v3 · encoding/json (stdlib) · net/http (stdlib) · archive/tar (stdlib)

---

## Dateistruktur

```
forgecrate/
├── cmd/forgecrate/
│   ├── main.go              # Einstiegspunkt, cobra root command
│   ├── init.go              # init subcommand
│   └── update.go            # update subcommand
├── internal/
│   ├── config/
│   │   ├── config.go        # .forgecrate.yaml lesen/schreiben
│   │   └── config_test.go
│   ├── github/
│   │   ├── client.go        # GitHub API: Tarball-Download
│   │   └── client_test.go
│   ├── compose/
│   │   ├── markdown.go      # CLAUDE.md/AGENTS.md Merge (GENERATED/CUSTOM Marker)
│   │   ├── markdown_test.go
│   │   ├── jsonmerge.go     # settings.json Deep-Merge
│   │   ├── jsonmerge_test.go
│   │   ├── skills.go        # .claude/commands/ Komposition
│   │   ├── skills_test.go
│   │   ├── compose.go       # Layer-Koordinator
│   │   └── compose_test.go
│   └── deploy/
│       ├── deploy.go        # Schreibt compositionierte Dateien ins Ziel-Repo
│       └── deploy_test.go
├── e2e/
│   └── e2e_test.go
├── base/
│   ├── CLAUDE.md
│   ├── AGENTS.md
│   ├── .claude/
│   │   ├── settings.json
│   │   └── commands/
│   └── hooks/
│       ├── prompt-submit.sh
│       └── pre-tool.sh
├── profiles/
│   ├── backend/CLAUDE.md
│   ├── frontend/CLAUDE.md
│   └── fullstack/CLAUDE.md
├── flavors/
│   ├── tdd/CLAUDE.md
│   ├── strict-review/CLAUDE.md
│   └── minimal/CLAUDE.md
├── assets/
│   └── banner.svg
├── docs/
│   ├── architecture.md
│   ├── flows.md
│   ├── layer-system.md
│   ├── hooks.md
│   ├── profiles-flavors.md
│   └── development.md
├── go.mod
├── go.sum
└── README.md
```

---

## Phase 1: Go Binary

### Task 1: Go-Modul + CLI-Skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/forgecrate/main.go`
- Create: `cmd/forgecrate/init.go`
- Create: `cmd/forgecrate/update.go`

- [ ] **Schritt 1: go.mod anlegen**

```bash
cd /Users/markus/repo/forgecrate
go mod init github.com/jmt-labs/forgecrate
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
```

- [ ] **Schritt 2: main.go schreiben**

```go
// cmd/forgecrate/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "forgecrate",
		Short: "Reproducible Claude Code configuration for Git repositories.",
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newUpdateCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Schritt 3: init.go Skeleton schreiben**

```go
// cmd/forgecrate/init.go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profile string
	var flavors []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("init: profile=%s flavors=%v\n", profile, flavors)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "backend", "Profil (backend|frontend|fullstack)")
	cmd.Flags().StringSliceVar(&flavors, "flavors", nil, "Flavors (tdd,strict-review,minimal)")
	return cmd
}
```

- [ ] **Schritt 4: update.go Skeleton schreiben**

```go
// cmd/forgecrate/update.go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Aktualisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("update: profile=%s\n", profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profil wechseln (optional)")
	return cmd
}
```

- [ ] **Schritt 5: Build prüfen**

```bash
go build ./cmd/forgecrate/
```

Erwartung: Binary `forgecrate` ohne Fehler.

- [ ] **Schritt 6: Committen**

```bash
git add go.mod go.sum cmd/
git commit -m "feat: add Go module and cobra CLI skeleton"
```

---

### Task 2: Config-Paket (`.forgecrate.yaml`)

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

```go
// internal/config/config_test.go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "strict-review"},
	}

	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Profile != "backend" {
		t.Errorf("Profile: got %q, want %q", got.Profile, "backend")
	}
	if len(got.Flavors) != 2 {
		t.Errorf("Flavors: got %d, want 2", len(got.Flavors))
	}
}

func TestReadMissing(t *testing.T) {
	_, err := config.Read("/nonexistent/.forgecrate.yaml")
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/config/...
```

Erwartung: `cannot find package`

- [ ] **Schritt 3: config.go implementieren**

```go
// internal/config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string   `yaml:"version"`
	Source  string   `yaml:"source"`
	Ref     string   `yaml:"ref"`
	Profile string   `yaml:"profile"`
	Flavors []string `yaml:"flavors"`
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
	defer f.Close()
	return yaml.NewEncoder(f).Encode(cfg)
}
```

- [ ] **Schritt 4: Test ausführen — muss bestehen**

```bash
go test ./internal/config/... -v
```

Erwartung: `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/config/
git commit -m "feat: add config package for .forgecrate.yaml"
```

---

### Task 3: GitHub-Client (Tarball-Download)

**Files:**
- Create: `internal/github/client.go`
- Create: `internal/github/client_test.go`

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/github/client_test.go
package github_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	gh "github.com/jmt-labs/forgecrate/internal/github"
)

func makeTarGz(files map[string]string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: "repo-prefix/" + name, Mode: 0644, Size: int64(len(content))}
		tw.WriteHeader(hdr)
		tw.Write([]byte(content))
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func TestDownloadAndExtract(t *testing.T) {
	tarball := makeTarGz(map[string]string{
		"base/CLAUDE.md": "# Base",
		"base/.claude/settings.json": `{"hooks":{}}`,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	dir := t.TempDir()

	if err := client.Download("markus", "forgecrate", "main", dir); err != nil {
		t.Fatalf("Download: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "base", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "# Base" {
		t.Errorf("got %q, want %q", string(content), "# Base")
	}
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/github/...
```

Erwartung: `cannot find package`

- [ ] **Schritt 3: client.go implementieren**

```go
// internal/github/client.go
package github

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	baseURL string
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func Default() *Client {
	return &Client{baseURL: "https://api.github.com"}
}

func (c *Client) Download(owner, repo, ref, destDir string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", c.baseURL, owner, repo, ref)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return extractTarGz(resp.Body, destDir)
}

func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		// Strip leading path component (GitHub adds "owner-repo-sha/" prefix)
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		rel := parts[1]
		if rel == "" {
			continue
		}

		dst := filepath.Join(destDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}
```

- [ ] **Schritt 4: Test ausführen — muss bestehen**

```bash
go test ./internal/github/... -v
```

Erwartung: `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/github/
git commit -m "feat: add github client with tarball download"
```

---

### Task 4: Markdown-Merge (CLAUDE.md / AGENTS.md)

**Files:**
- Create: `internal/compose/markdown.go`
- Create: `internal/compose/markdown_test.go`

Das Format der compositionierten Datei:

```markdown
<!-- GENERATED:BEGIN -->
[Generierter Inhalt — wird bei update überschrieben]
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
[Eigener Inhalt — wird nie überschrieben]
<!-- CUSTOM:END -->
```

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/compose/markdown_test.go
package compose_test

import (
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestMergeMarkdownInit(t *testing.T) {
	layers := []string{"# Base\n\nBase content.", "## Profile\n\nProfile content."}
	result := compose.MergeMarkdown(layers, "")

	want := "<!-- GENERATED:BEGIN -->\n# Base\n\nBase content.\n\n## Profile\n\nProfile content.\n<!-- GENERATED:END -->\n\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"
	if result != want {
		t.Errorf("got:\n%q\nwant:\n%q", result, want)
	}
}

func TestMergeMarkdownPreservesCustom(t *testing.T) {
	existing := "<!-- GENERATED:BEGIN -->\n# Old\n<!-- GENERATED:END -->\n\n<!-- CUSTOM:BEGIN -->\n# My custom section\n<!-- CUSTOM:END -->\n"
	layers := []string{"# New Base"}
	result := compose.MergeMarkdown(layers, existing)

	if !contains(result, "# My custom section") {
		t.Error("custom section was lost")
	}
	if !contains(result, "# New Base") {
		t.Error("new generated content missing")
	}
}

func TestMergeMarkdownNoExistingMarkers(t *testing.T) {
	existing := "# Handwritten file\n\nNo markers here."
	layers := []string{"# Base"}
	result := compose.MergeMarkdown(layers, existing)

	if !contains(result, "# Handwritten file") {
		t.Error("existing content without markers was lost")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/compose/... -run TestMergeMarkdown
```

Erwartung: `cannot find package`

- [ ] **Schritt 3: markdown.go implementieren**

```go
// internal/compose/markdown.go
package compose

import (
	"strings"
)

const (
	generatedBegin = "<!-- GENERATED:BEGIN -->"
	generatedEnd   = "<!-- GENERATED:END -->"
	customBegin    = "<!-- CUSTOM:BEGIN -->"
	customEnd      = "<!-- CUSTOM:END -->"
)

// MergeMarkdown compositioniert mehrere Markdown-Layer zu einem String.
// existing ist der aktuelle Dateiinhalt (leer bei init).
func MergeMarkdown(layers []string, existing string) string {
	generated := strings.Join(layers, "\n\n")
	custom := extractCustom(existing)

	var b strings.Builder
	b.WriteString(generatedBegin + "\n")
	b.WriteString(generated + "\n")
	b.WriteString(generatedEnd + "\n")
	b.WriteString("\n")
	b.WriteString(customBegin + "\n")
	b.WriteString(custom)
	b.WriteString(customEnd + "\n")
	return b.String()
}

func extractCustom(existing string) string {
	start := strings.Index(existing, customBegin)
	end := strings.Index(existing, customEnd)

	if start == -1 || end == -1 {
		// Keine Marker: existierende Datei als Custom behandeln
		if strings.TrimSpace(existing) != "" {
			return existing + "\n"
		}
		return ""
	}

	content := existing[start+len(customBegin) : end]
	return strings.TrimLeft(content, "\n")
}
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/compose/... -run TestMergeMarkdown -v
```

Erwartung: alle 3 Tests `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/compose/markdown.go internal/compose/markdown_test.go
git commit -m "feat: add markdown merge with GENERATED/CUSTOM markers"
```

---

### Task 5: JSON Deep-Merge (`settings.json`)

**Files:**
- Create: `internal/compose/jsonmerge.go`
- Create: `internal/compose/jsonmerge_test.go`

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/compose/jsonmerge_test.go
package compose_test

import (
	"encoding/json"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestDeepMergeJSON(t *testing.T) {
	base := `{"hooks":{"UserPromptSubmit":[{"matcher":"","hooks":[{"type":"command","command":"bash a.sh"}]}]},"permissions":{"allow":["Bash"]}}`
	override := `{"permissions":{"allow":["Bash","Edit"]},"model":"claude-opus-4-7"}`

	result, err := compose.DeepMergeJSON(base, override)
	if err != nil {
		t.Fatalf("DeepMergeJSON: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}

	// Override-Wert gewinnt
	if m["model"] != "claude-opus-4-7" {
		t.Errorf("model: got %v", m["model"])
	}

	// Arrays aus override ersetzen base (nicht append)
	perms := m["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	if len(allow) != 2 {
		t.Errorf("allow: got %d elements, want 2", len(allow))
	}

	// Base-Keys bleiben erhalten wenn override sie nicht hat
	hooks := m["hooks"].(map[string]any)
	if hooks["UserPromptSubmit"] == nil {
		t.Error("hooks.UserPromptSubmit missing from merge result")
	}
}

func TestDeepMergeJSONEmpty(t *testing.T) {
	result, err := compose.DeepMergeJSON(`{"a":1}`, `{}`)
	if err != nil {
		t.Fatalf("DeepMergeJSON: %v", err)
	}
	var m map[string]any
	json.Unmarshal([]byte(result), &m)
	if m["a"] != float64(1) {
		t.Errorf("a: got %v", m["a"])
	}
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/compose/... -run TestDeepMergeJSON
```

Erwartung: `undefined: compose.DeepMergeJSON`

- [ ] **Schritt 3: jsonmerge.go implementieren**

```go
// internal/compose/jsonmerge.go
package compose

import (
	"encoding/json"
	"fmt"
)

// DeepMergeJSON merged zwei JSON-Objekte. Override-Werte gewinnen.
// Arrays werden ersetzt (nicht gemergt). Objekte werden rekursiv gemergt.
func DeepMergeJSON(base, override string) (string, error) {
	var baseMap, overrideMap map[string]any

	if err := json.Unmarshal([]byte(base), &baseMap); err != nil {
		return "", fmt.Errorf("base JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(override), &overrideMap); err != nil {
		return "", fmt.Errorf("override JSON: %w", err)
	}

	merged := deepMerge(baseMap, overrideMap)
	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func deepMerge(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		if baseVal, ok := result[k]; ok {
			baseMap, baseIsMap := baseVal.(map[string]any)
			overrideMap, overrideIsMap := v.(map[string]any)
			if baseIsMap && overrideIsMap {
				result[k] = deepMerge(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/compose/... -run TestDeepMergeJSON -v
```

Erwartung: `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/compose/jsonmerge.go internal/compose/jsonmerge_test.go
git commit -m "feat: add deep JSON merge for settings.json"
```

---

### Task 6: Skills-Komposition (`.claude/commands/`)

**Files:**
- Create: `internal/compose/skills.go`
- Create: `internal/compose/skills_test.go`

- [ ] **Schritt 1: Failing Tests schreiben**

```go
// internal/compose/skills_test.go
package compose_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestMergeSkills(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Source: base + profile skills
	os.WriteFile(filepath.Join(src, "base.md"), []byte("# Base Skill"), 0644)
	os.WriteFile(filepath.Join(src, "shared.md"), []byte("# Shared Base"), 0644)

	// Override im Ziel-Repo (darf nicht überschrieben werden)
	overridesDir := filepath.Join(dst, "overrides")
	os.MkdirAll(overridesDir, 0755)
	os.WriteFile(filepath.Join(overridesDir, "shared.md"), []byte("# My Override"), 0644)

	if err := compose.MergeSkills([]string{src}, dst); err != nil {
		t.Fatalf("MergeSkills: %v", err)
	}

	// Base-Skill wurde kopiert
	content, err := os.ReadFile(filepath.Join(dst, "base.md"))
	if err != nil {
		t.Fatalf("base.md not found: %v", err)
	}
	if string(content) != "# Base Skill" {
		t.Errorf("base.md: got %q", content)
	}

	// Override wurde nicht überschrieben
	override, _ := os.ReadFile(filepath.Join(overridesDir, "shared.md"))
	if string(override) != "# My Override" {
		t.Errorf("override was overwritten: got %q", override)
	}
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/compose/... -run TestMergeSkills
```

Erwartung: `undefined: compose.MergeSkills`

- [ ] **Schritt 3: skills.go implementieren**

```go
// internal/compose/skills.go
package compose

import (
	"io"
	"os"
	"path/filepath"
)

// MergeSkills kopiert Skill-Dateien aus allen srcDirs in destDir.
// Spätere srcDirs überschreiben frühere (Layer-Precedence).
// Dateien unter destDir/overrides/ werden nie überschrieben.
func MergeSkills(srcDirs []string, destDir string) error {
	for _, src := range srcDirs {
		if err := copyDir(src, destDir); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, rel)

		// Overrides nie anfassen
		if isUnderOverrides(dst, dstPath) {
			return nil
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return copyFile(path, dstPath)
	})
}

func isUnderOverrides(destDir, path string) bool {
	overridesDir := filepath.Join(destDir, "overrides")
	rel, err := filepath.Rel(overridesDir, path)
	if err != nil {
		return false
	}
	return len(rel) > 0 && rel[0] != '.'
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/compose/... -run TestMergeSkills -v
```

Erwartung: `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/compose/skills.go internal/compose/skills_test.go
git commit -m "feat: add skills composition with override protection"
```

---

### Task 7: Layer-Koordinator (`compose.go`)

**Files:**
- Create: `internal/compose/compose.go`
- Create: `internal/compose/compose_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

```go
// internal/compose/compose_test.go
package compose_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestCompose(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Minimale Source-Struktur
	writeFile(t, src, "base/CLAUDE.md", "# Base Claude")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":["Bash"]}}`)
	writeFile(t, src, "base/.claude/commands/base-skill.md", "# Base Skill")
	writeFile(t, src, "profiles/backend/CLAUDE.md", "## Backend Profile")
	writeFile(t, src, "flavors/tdd/CLAUDE.md", "## TDD Flavor")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "backend",
		Flavors:   []string{"tdd"},
	}

	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// CLAUDE.md wurde compositioniert
	claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md missing: %v", err)
	}
	for _, want := range []string{"# Base Claude", "## Backend Profile", "## TDD Flavor"} {
		if !containsStr(string(claudeMD), want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}

	// settings.json vorhanden
	if _, err := os.Stat(filepath.Join(dst, ".claude", "settings.json")); err != nil {
		t.Errorf("settings.json missing: %v", err)
	}

	// Skill kopiert
	if _, err := os.Stat(filepath.Join(dst, ".claude", "commands", "base-skill.md")); err != nil {
		t.Errorf("base-skill.md missing: %v", err)
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/compose/... -run TestCompose
```

Erwartung: `undefined: compose.Run`

- [ ] **Schritt 3: compose.go implementieren**

```go
// internal/compose/compose.go
package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Request struct {
	SourceDir string
	DestDir   string
	Profile   string
	Flavors   []string
}

func Run(req Request) error {
	if err := composeMarkdown(req, "CLAUDE.md"); err != nil {
		return fmt.Errorf("CLAUDE.md: %w", err)
	}
	if err := composeMarkdown(req, "AGENTS.md"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md: %w", err)
	}
	if err := composeJSON(req); err != nil {
		return fmt.Errorf("settings.json: %w", err)
	}
	if err := composeSkills(req); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
	return nil
}

func composeMarkdown(req Request, filename string) error {
	layers := collectMarkdownLayers(req, filename)
	if len(layers) == 0 {
		return os.ErrNotExist
	}

	existing := ""
	if data, err := os.ReadFile(filepath.Join(req.DestDir, filename)); err == nil {
		existing = string(data)
	}

	result := MergeMarkdown(layers, existing)
	return os.WriteFile(filepath.Join(req.DestDir, filename), []byte(result), 0644)
}

func collectMarkdownLayers(req Request, filename string) []string {
	var layers []string
	candidates := []string{
		filepath.Join(req.SourceDir, "base", filename),
		filepath.Join(req.SourceDir, "profiles", req.Profile, filename),
	}
	for _, flavor := range req.Flavors {
		candidates = append(candidates, filepath.Join(req.SourceDir, "flavors", flavor, filename))
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			layers = append(layers, string(data))
		}
	}
	return layers
}

func composeJSON(req Request) error {
	basePath := filepath.Join(req.SourceDir, "base", ".claude", "settings.json")
	data, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}
	merged := string(data)

	// Profile settings.json (falls vorhanden)
	profilePath := filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "settings.json")
	if override, err := os.ReadFile(profilePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return err
		}
	}

	// Override aus Ziel-Repo (falls vorhanden)
	overridePath := filepath.Join(req.DestDir, ".claude", "overrides", "settings.override.json")
	if override, err := os.ReadFile(overridePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return err
		}
	}

	// Validierung
	var v any
	if err := json.Unmarshal([]byte(merged), &v); err != nil {
		return fmt.Errorf("merged JSON invalid: %w", err)
	}

	dst := filepath.Join(req.DestDir, ".claude", "settings.json")
	os.MkdirAll(filepath.Dir(dst), 0755)
	return os.WriteFile(dst, []byte(merged), 0644)
}

func composeSkills(req Request) error {
	skillsDest := filepath.Join(req.DestDir, ".claude", "commands")
	os.MkdirAll(skillsDest, 0755)

	srcDirs := []string{
		filepath.Join(req.SourceDir, "base", ".claude", "commands"),
		filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "commands"),
	}
	for _, flavor := range req.Flavors {
		srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "flavors", flavor, ".claude", "commands"))
	}

	// Nur existierende Dirs übergeben
	var existing []string
	for _, d := range srcDirs {
		if _, err := os.Stat(d); err == nil {
			existing = append(existing, d)
		}
	}

	return MergeSkills(existing, skillsDest)
}
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/compose/... -v
```

Erwartung: alle Tests `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/compose/compose.go internal/compose/compose_test.go
git commit -m "feat: add layer composer coordinating markdown/json/skills"
```

---

### Task 8: Deploy-Paket

**Files:**
- Create: `internal/deploy/deploy.go`
- Create: `internal/deploy/deploy_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

```go
// internal/deploy/deploy_test.go
package deploy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

func TestDeploy(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Minimale Source-Struktur
	writeFile(t, src, "base/CLAUDE.md", "# Base")
	writeFile(t, src, "base/.claude/settings.json", `{"hooks":{}}`)
	writeFile(t, src, "base/hooks/prompt-submit.sh", "#!/bin/bash\necho ok")
	writeFile(t, src, "base/hooks/pre-tool.sh", "#!/bin/bash\necho ok")

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// .forgecrate.yaml geschrieben
	if _, err := os.Stat(filepath.Join(dst, ".forgecrate.yaml")); err != nil {
		t.Errorf(".forgecrate.yaml missing")
	}

	// Hooks kopiert
	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "prompt-submit.sh")); err != nil {
		t.Errorf("prompt-submit.sh missing")
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)
}
```

- [ ] **Schritt 2: Test ausführen — muss fehlschlagen**

```bash
go test ./internal/deploy/...
```

Erwartung: `cannot find package`

- [ ] **Schritt 3: deploy.go implementieren**

```go
// internal/deploy/deploy.go
package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jmt-labs/forgecrate/internal/compose"
	"github.com/jmt-labs/forgecrate/internal/config"
)

// Run compositioniert alle Layer und schreibt das Ergebnis ins destDir.
func Run(sourceDir, destDir string, cfg config.Config) error {
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

	cfgPath := filepath.Join(destDir, ".forgecrate.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func copyHooks(src, dst string) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil
	}

	dstHooks := filepath.Join(dst, ".claude", "hooks")
	os.MkdirAll(dstHooks, 0755)

	return filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(hooksDir, path)
		dstPath := filepath.Join(dstHooks, rel)

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/deploy/... -v
```

Erwartung: `PASS`

- [ ] **Schritt 5: Committen**

```bash
git add internal/deploy/
git commit -m "feat: add deploy package wiring compose and config write"
```

---

### Task 9: `init`- und `update`-Commands verdrahten

**Files:**
- Modify: `cmd/forgecrate/init.go`
- Modify: `cmd/forgecrate/update.go`

- [ ] **Schritt 1: init.go mit echter Logik überschreiben**

```go
// cmd/forgecrate/init.go
package main

import (
	"fmt"
	"os"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profile string
	var flavors []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfgPath := cwd + "/.forgecrate.yaml"
			cfg, err := config.Read(cfgPath)
			if os.IsNotExist(err) {
				cfg = config.Config{
					Version: "1.0",
					Source:  "github.com/jmt-labs/forgecrate",
					Ref:     "main",
					Profile: profile,
					Flavors: flavors,
				}
			} else if err != nil {
				return err
			}

			if profile != "" {
				cfg.Profile = profile
			}
			if len(flavors) > 0 {
				cfg.Flavors = flavors
			}

			owner, repo := "markus", "forgecrate"
			fmt.Printf("Fetching %s/%s@%s ...\n", owner, repo, cfg.Ref)

			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(srcDir)

			client := gh.Default()
			if err := client.Download(owner, repo, cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("Deploying profile=%s flavors=%v ...\n", cfg.Profile, cfg.Flavors)
			if err := deploy.Run(srcDir, cwd, cfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "backend", "Profil (backend|frontend|fullstack)")
	cmd.Flags().StringSliceVar(&flavors, "flavors", nil, "Flavors (tdd,strict-review,minimal)")
	return cmd
}
```

- [ ] **Schritt 2: update.go mit echter Logik überschreiben**

```go
// cmd/forgecrate/update.go
package main

import (
	"fmt"
	"os"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
	gh "github.com/jmt-labs/forgecrate/internal/github"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Aktualisiert forgecrate im aktuellen Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := config.Read(cwd + "/.forgecrate.yaml")
			if os.IsNotExist(err) {
				return fmt.Errorf(".forgecrate.yaml nicht gefunden — erst 'forgecrate init' ausführen")
			} else if err != nil {
				return err
			}

			if profile != "" {
				cfg.Profile = profile
			}

			owner, repo := "markus", "forgecrate"
			fmt.Printf("Fetching %s/%s@%s ...\n", owner, repo, cfg.Ref)

			srcDir, err := os.MkdirTemp("", "forgecrate-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(srcDir)

			client := gh.Default()
			if err := client.Download(owner, repo, cfg.Ref, srcDir); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("Updating profile=%s flavors=%v ...\n", cfg.Profile, cfg.Flavors)
			if err := deploy.Run(srcDir, cwd, cfg); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Profil wechseln (optional)")
	return cmd
}
```

- [ ] **Schritt 3: Build + Smoketest**

```bash
go build ./cmd/forgecrate/
./forgecrate --help
./forgecrate init --help
./forgecrate update --help
```

Erwartung: Hilfe-Text ohne Fehler.

- [ ] **Schritt 4: Committen**

```bash
git add cmd/forgecrate/
git commit -m "feat: wire init and update commands with github+deploy"
```

---

## Phase 2: Content-Dateien

### Task 10: Base-Content

**Files:**
- Create: `base/CLAUDE.md`
- Create: `base/AGENTS.md`
- Create: `base/.claude/settings.json`
- Create: `base/hooks/prompt-submit.sh`
- Create: `base/hooks/pre-tool.sh`

- [ ] **Schritt 1: `base/CLAUDE.md` schreiben**

```markdown
<!-- GENERATED:BEGIN -->
# Claude-Konfiguration

Dieses Repository verwendet ein reproduzierbares forgecrate.
Die generierten Abschnitte dieser Datei werden bei `forgecrate update` überschrieben.
Eigene Anpassungen gehören in den CUSTOM-Abschnitt.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer minimal und zielgerichtet
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->
```

- [ ] **Schritt 2: `base/AGENTS.md` schreiben**

```markdown
<!-- GENERATED:BEGIN -->
# Agent-Konfiguration

Gilt für alle Agenten (Codex, Claude Code, etc.) die in diesem Repo arbeiten.

## Pflichten

- Vor jeder Code-Änderung den relevanten Kontext vollständig lesen
- Tests schreiben bevor Implementierung
- Commits nach jeder abgeschlossenen Aufgabe
- Keine globalen Konfigurationen verändern
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->
```

- [ ] **Schritt 3: `base/.claude/settings.json` schreiben**

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bash .claude/hooks/prompt-submit.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash|Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash .claude/hooks/pre-tool.sh"
          }
        ]
      }
    ]
  },
  "permissions": {
    "allow": [],
    "deny": []
  }
}
```

- [ ] **Schritt 4: `base/hooks/prompt-submit.sh` schreiben**

```bash
#!/usr/bin/env bash
# Erinnerung an Pflicht-Skills — wird bei jeder User-Nachricht ausgegeben.
# Schlank halten: nur wenige Zeilen, vollständig cached nach erster Ausführung.

PROFILE=$(grep 'profile:' .forgecrate.yaml 2>/dev/null | awk '{print $2}')
FLAVORS=$(grep -A5 'flavors:' .forgecrate.yaml 2>/dev/null | grep '  -' | awk '{print $2}' | tr '\n' ',' | sed 's/,$//')

echo "## forgecrate — Aktive Konfiguration"
echo "Profil: ${PROFILE:-unbekannt} | Flavors: ${FLAVORS:-keine}"
echo ""
echo "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs"
```

- [ ] **Schritt 5: `base/hooks/pre-tool.sh` schreiben**

```bash
#!/usr/bin/env bash
# Wird vor Bash/Edit/Write aufgerufen.
# Gibt eine Warnung aus wenn keine Session-Notiz über einen aufgerufenen Skill existiert.
# Claude sieht diese Ausgabe als Kontext.

TOOL="${CLAUDE_TOOL_NAME:-}"

case "$TOOL" in
  Edit|Write)
    echo "## Pre-Tool Check"
    echo "Du verwendest $TOOL. Stelle sicher:"
    echo "- superpowers:brainstorming wurde für neue Features aufgerufen"
    echo "- superpowers:test-driven-development wurde vor der Implementierung aufgerufen"
    ;;
  Bash)
    echo "## Pre-Tool Check"
    echo "Du verwendest Bash. Stelle sicher dass destruktive Aktionen mit dem User abgestimmt sind."
    ;;
esac
```

- [ ] **Schritt 6: Committen**

```bash
git add base/
git commit -m "feat: add base content (CLAUDE.md, AGENTS.md, settings.json, hooks)"
```

---

### Task 11: Profile und Flavors

**Files:**
- Create: `profiles/backend/CLAUDE.md`
- Create: `profiles/frontend/CLAUDE.md`
- Create: `profiles/fullstack/CLAUDE.md`
- Create: `flavors/tdd/CLAUDE.md`
- Create: `flavors/strict-review/CLAUDE.md`
- Create: `flavors/minimal/CLAUDE.md`

- [ ] **Schritt 1: `profiles/backend/CLAUDE.md`**

```markdown
## Backend-Profil

- API-Design: REST-First, klare Fehlercodes, keine unnötige Abstraktion
- Datenbankzugriffe: typsicher, keine Raw-Queries ohne Parametrisierung
- Tests: Integrationstests bevorzugt gegenüber reinen Unit-Tests mit Mocks
- Kein ORM-Magic: explizite Queries sind verständlicher
```

- [ ] **Schritt 2: `profiles/frontend/CLAUDE.md`**

```markdown
## Frontend-Profil

- Komponenten: klein, fokussiert, eine Verantwortlichkeit
- State: lokal wenn möglich, global nur wenn nötig
- Kein CSS-in-JS ohne explizite Anforderung
- Barrierefreiheit: semantisches HTML, ARIA-Attribute wo nötig
- Tests: Behavior-Tests (was der Nutzer sieht), keine Implementierungsdetails
```

- [ ] **Schritt 3: `profiles/fullstack/CLAUDE.md`**

```markdown
## Fullstack-Profil

Kombiniert Backend- und Frontend-Anforderungen.

- API-Kontrakte explizit definieren bevor Implementierung auf beiden Seiten
- Shared Types: einmal definieren, in beiden Schichten nutzen
- End-to-End-Tests für kritische User-Flows
```

- [ ] **Schritt 4: `flavors/tdd/CLAUDE.md`**

```markdown
## TDD-Flavor

- Test schreiben → ausführen (muss fehlschlagen) → implementieren → ausführen (muss bestehen) → committen
- Kein Produktionscode ohne vorherigen Test
- Test-Namen beschreiben Verhalten, nicht Implementierung
- Mocks nur an Systemgrenzen (externe APIs, Datenbanken)
```

- [ ] **Schritt 5: `flavors/strict-review/CLAUDE.md`**

```markdown
## Strict-Review-Flavor

- Vor jedem Commit: `superpowers:requesting-code-review` aufrufen
- Keine direkten Commits auf main/master
- PR-Beschreibung enthält: Was, Warum, Wie getestet
- Breaking Changes werden explizit kommuniziert
```

- [ ] **Schritt 6: `flavors/minimal/CLAUDE.md`**

```markdown
## Minimal-Flavor

- Keine Pflicht-Skills außer `verification-before-completion`
- Kein TDD-Zwang — Tests wo sinnvoll
- Hooks aktiv aber keine Blockierung
```

- [ ] **Schritt 7: Committen**

```bash
git add profiles/ flavors/
git commit -m "feat: add profiles (backend/frontend/fullstack) and flavors (tdd/strict-review/minimal)"
```

---

## Phase 3: Tests & Dokumentation

### Task 12: E2E-Tests

**Files:**
- Create: `e2e/e2e_test.go`

- [ ] **Schritt 1: E2E-Test schreiben**

```go
// e2e/e2e_test.go
package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

// localSource gibt den Pfad zum lokalen forgecrate Repo zurück.
// E2E-Tests laufen gegen lokale Quellen, nicht gegen GitHub.
func localSource(t *testing.T) string {
	t.Helper()
	// e2e/ liegt ein Level unterhalb des Repo-Root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(wd) // Repo-Root
}

func TestInitCreatesExpectedFiles(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	requiredFiles := []string{
		".forgecrate.yaml",
		"CLAUDE.md",
		"AGENTS.md",
		".claude/settings.json",
		".claude/hooks/prompt-submit.sh",
		".claude/hooks/pre-tool.sh",
	}

	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(dst, f)); err != nil {
			t.Errorf("missing: %s", f)
		}
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// Custom-Section manuell setzen
	claudeMD := filepath.Join(dst, "CLAUDE.md")
	content, _ := os.ReadFile(claudeMD)
	customized := strings.Replace(string(content),
		"<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->",
		"<!-- CUSTOM:BEGIN -->\n# Mein Custom\n<!-- CUSTOM:END -->", 1)
	os.WriteFile(claudeMD, []byte(customized), 0644)

	// Zweiter Lauf
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("second run: %v", err)
	}

	after, _ := os.ReadFile(claudeMD)
	if !strings.Contains(string(after), "# Mein Custom") {
		t.Error("custom section was lost on second run")
	}
}

func TestUpdatePreservesOverrides(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	// Init
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Override-Skill anlegen
	overridesDir := filepath.Join(dst, ".claude", "commands", "overrides")
	os.MkdirAll(overridesDir, 0755)
	overrideSkill := filepath.Join(overridesDir, "my-skill.md")
	os.WriteFile(overrideSkill, []byte("# My Skill"), 0644)

	// Update
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Override erhalten?
	content, err := os.ReadFile(overrideSkill)
	if err != nil {
		t.Fatalf("override skill missing after update")
	}
	if string(content) != "# My Skill" {
		t.Errorf("override was modified: %q", content)
	}
}

func TestProfileSwitch(t *testing.T) {
	dst := t.TempDir()

	cfg := config.Config{Version: "1.0", Source: "github.com/jmt-labs/forgecrate", Ref: "main", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("init backend: %v", err)
	}

	cfg.Profile = "frontend"
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("update to frontend: %v", err)
	}

	yamlContent, _ := os.ReadFile(filepath.Join(dst, ".forgecrate.yaml"))
	if !strings.Contains(string(yamlContent), "frontend") {
		t.Error("profile not updated in .forgecrate.yaml")
	}

	claudeMD, _ := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if !strings.Contains(string(claudeMD), "Frontend-Profil") {
		t.Error("frontend profile content missing in CLAUDE.md")
	}
}
```

- [ ] **Schritt 2: E2E-Tests ausführen**

```bash
cd /Users/markus/repo/forgecrate
go test ./e2e/... -v
```

Erwartung: alle 4 Tests `PASS`

- [ ] **Schritt 3: Vollständige Test-Suite ausführen**

```bash
go test ./...
```

Erwartung: alle Tests `PASS`, kein `FAIL`

- [ ] **Schritt 4: Committen**

```bash
git add e2e/
git commit -m "feat: add e2e tests for init/update/idempotency/profile-switch"
```

---

### Task 13: SVG-Banner + Technische Dokumentation

**Files:**
- Create: `assets/banner.svg`
- Create: `docs/architecture.md`
- Create: `docs/flows.md`
- Create: `docs/layer-system.md`
- Create: `docs/hooks.md`
- Create: `docs/profiles-flavors.md`
- Create: `docs/development.md`

- [ ] **Schritt 1: `assets/banner.svg` erstellen**

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 120" width="800" height="120">
  <rect width="800" height="120" fill="#0d1117"/>
  <text x="400" y="52" font-family="monospace" font-size="28" font-weight="bold"
        fill="#e6edf3" text-anchor="middle" letter-spacing="2">forgecrate</text>
  <text x="400" y="82" font-family="monospace" font-size="14"
        fill="#8b949e" text-anchor="middle">Reproducible Claude Code configuration for Git repositories</text>
  <text x="400" y="106" font-family="monospace" font-size="11"
        fill="#6e7681" text-anchor="middle">Go Binary · Layer System · Hooks · GitHub</text>
</svg>
```

- [ ] **Schritt 2: `docs/architecture.md` schreiben**

```markdown
# Architektur

## Komponenten

```
┌─────────────────────────────────────────────────────┐
│                   forgecrate Repo                  │
│  base/ · profiles/ · flavors/ · cmd/forgecrate/   │
└───────────────────────┬─────────────────────────────┘
                        │ GitHub API (tarball)
                        ▼
              ┌─────────────────┐
              │  forgecrate   │  ← globales Go-Binary
              │    Binary       │
              └────────┬────────┘
                       │ compose + deploy
                       ▼
        ┌──────────────────────────────┐
        │          Ziel-Repo           │
        │  CLAUDE.md  AGENTS.md        │
        │  .claude/settings.json       │
        │  .claude/commands/           │
        │  .claude/hooks/              │
        │  .forgecrate.yaml          │
        └──────────────────────────────┘
```

## Layer-System

```
Layer 1: base/          → immer aktiv
Layer 2: profiles/<p>/  → eines wählbar
Layer 2: flavors/<f>/   → mehrere kombinierbar
Layer 3: overrides/     → lokal, nie überschrieben
```
```

- [ ] **Schritt 3: `docs/flows.md` schreiben**

```markdown
# Abläufe

## init-Flow

```
forgecrate init --profile backend --flavors tdd
        │
        ├── .forgecrate.yaml lesen (falls vorhanden)
        ├── GitHub tarball downloaden → tmpDir
        ├── Layer compositionieren: base → profile → flavors
        │       ├── CLAUDE.md: MergeMarkdown(layers, existing)
        │       ├── AGENTS.md: MergeMarkdown(layers, existing)
        │       ├── settings.json: DeepMergeJSON(base, profile, overrides)
        │       └── commands/: MergeSkills(srcDirs, dest)
        ├── Hooks nach .claude/hooks/ kopieren
        ├── .forgecrate.yaml schreiben
        └── Done.
```

## update-Flow

```
forgecrate update [--profile <p>]
        │
        ├── .forgecrate.yaml lesen (Fehler wenn nicht vorhanden)
        ├── Profile überschreiben wenn --profile angegeben
        ├── GitHub tarball downloaden → tmpDir
        ├── Layer rekompositionieren (overrides/ unangetastet)
        └── Done.
```

## Enforcement-Flow

```
User schreibt Prompt
        │
        ├── UserPromptSubmit-Hook: prompt-submit.sh
        │       └── Gibt Profil + Pflicht-Skills aus
        │
        ├── Claude liest Prompt + CLAUDE.md-Pflicht-Skills-Tabelle
        ├── Claude ruft relevanten Skill auf
        │
        ├── PreToolUse-Hook (vor Bash/Edit/Write): pre-tool.sh
        │       └── Gibt kontextabhängige Erinnerung aus
        │
        └── Tool wird ausgeführt
```
```

- [ ] **Schritt 4: `docs/layer-system.md` schreiben**

```markdown
# Layer-System

## Drei Ebenen

| Layer | Quelle | Precedence |
|---|---|---|
| base | `base/` | Niedrigste — immer aktiv |
| profile | `profiles/<name>/` | Überschreibt base |
| flavor | `flavors/<name>/` | Überschreibt profile |
| override | `.claude/commands/overrides/` (Ziel-Repo) | Höchste — nie überschrieben |

## Kompositions-Regeln

### CLAUDE.md / AGENTS.md

Inhalte werden aneinandergehängt (base → profile → flavors).
Das Ergebnis landet im GENERATED-Block. Der CUSTOM-Block bleibt immer erhalten.

### settings.json

Deep JSON Merge. Objekte werden rekursiv gemergt. Arrays werden ersetzt.
`settings.override.json` im Ziel-Repo hat höchste Priorität.

### .claude/commands/

Alle Skill-Dateien werden additiv kopiert. Spätere Layer überschreiben frühere
gleichnamige Dateien. Dateien unter `overrides/` werden nie angefasst.

## Beispiel

```
base/CLAUDE.md          → "# Base..."
profiles/backend/CLAUDE.md → "## Backend..."
flavors/tdd/CLAUDE.md   → "## TDD..."

Ergebnis CLAUDE.md:
  <!-- GENERATED:BEGIN -->
  # Base...

  ## Backend...

  ## TDD...
  <!-- GENERATED:END -->

  <!-- CUSTOM:BEGIN -->
  [User-Inhalt]
  <!-- CUSTOM:END -->
```
```

- [ ] **Schritt 5: `docs/hooks.md` und `docs/profiles-flavors.md` schreiben**

`docs/hooks.md`:
```markdown
# Hooks

## UserPromptSubmit — `prompt-submit.sh`

Wird bei jeder User-Nachricht ausgeführt.

**Output:** Aktives Profil + Pflicht-Skill-Liste (wenige Zeilen, gecacht).

## PreToolUse — `pre-tool.sh`

Wird vor `Bash`, `Edit`, `Write` ausgeführt.

**Input:** `$CLAUDE_TOOL_NAME` (Tool-Name)

**Output:** Kontextabhängige Erinnerung an relevante Pflicht-Skills.

## Override

Hooks können in `overrides/settings.override.json` ergänzt oder ersetzt werden.
```

`docs/profiles-flavors.md`:
```markdown
# Profile und Flavors

## Profile (eines wählbar)

| Profil | Fokus |
|---|---|
| `backend` | API, Datenbank, Integrationstests |
| `frontend` | Komponenten, State, Barrierefreiheit |
| `fullstack` | Kombination beider, shared Types, E2E |

## Flavors (mehrere kombinierbar)

| Flavor | Fokus |
|---|---|
| `tdd` | Test-First, kein Produktionscode ohne Test |
| `strict-review` | Pflicht-Review vor jedem Commit |
| `minimal` | Nur Basis-Enforcement |
```

`docs/development.md`:
```markdown
# Entwicklung

## Voraussetzungen

Go 1.22+

## Tests ausführen

```bash
go test ./...        # Unit-Tests
go test ./e2e/...    # E2E-Tests (gegen lokales Repo)
```

## Binary bauen

```bash
go build ./cmd/forgecrate/
```

## Neues Profil hinzufügen

1. `profiles/<name>/CLAUDE.md` anlegen
2. Optional: `profiles/<name>/.claude/settings.json` für Profil-spezifische Settings
3. Optional: `profiles/<name>/.claude/commands/` für Profil-spezifische Skills
```

- [ ] **Schritt 6: Committen**

```bash
git add assets/ docs/
git commit -m "feat: add SVG banner and technical documentation"
```

---

### Task 14: README.md (forgedeck-Stil)

**Files:**
- Create: `README.md`

- [ ] **Schritt 1: README.md schreiben**

```markdown
<div align="center">
  <img src="assets/banner.svg" alt="forgecrate — Reproducible forgecrate" width="100%">
</div>

# forgecrate

forgecrate deployt ein reproduzierbares forgecrate in beliebige Repos. Ein globales Go-Binary holt Konfiguration, Skills und Hooks von GitHub und compositioniert sie per Layer-System ins Ziel-Repo.

Stack: Go Binary · GitHub API · Layer-System · Hooks · Skills

---

## Quick Start

Voraussetzungen: Go 1.22+

```sh
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest

# Im Ziel-Repo:
forgecrate init --profile backend --flavors tdd
```

Danach enthält das Repo:

```
CLAUDE.md · AGENTS.md · .claude/settings.json · .claude/commands/ · .claude/hooks/
```

Aktualisieren:

```sh
forgecrate update
```

Profil wechseln:

```sh
forgecrate update --profile fullstack
```

---

## Dokumentation

| Thema | Dokument |
|---|---|
| Architektur und Komponenten | [docs/architecture.md](docs/architecture.md) |
| Abläufe (init, update, enforcement) | [docs/flows.md](docs/flows.md) |
| Layer-System | [docs/layer-system.md](docs/layer-system.md) |
| Hooks | [docs/hooks.md](docs/hooks.md) |
| Profile und Flavors | [docs/profiles-flavors.md](docs/profiles-flavors.md) |
| Entwicklung und Tests | [docs/development.md](docs/development.md) |

---

## Komponenten

| Pfad | Zweck |
|---|---|
| `base/` | Basis-Layer — immer deployt |
| `profiles/` | Profil-Layer — eines wählbar |
| `flavors/` | Flavor-Layer — mehrere kombinierbar |
| `cmd/forgecrate/` | Go-Binary (init, update) |
| `internal/compose/` | Markdown-, JSON- und Skills-Merge-Logik |
| `internal/github/` | GitHub API Client |
| `internal/config/` | `.forgecrate.yaml` Lesen/Schreiben |
| `internal/deploy/` | Deployment-Koordination |
| `e2e/` | End-to-End-Tests |

---

## Anpassung

Lokale Overrides werden nie überschrieben:

```
.claude/commands/overrides/   # eigene Skills
.claude/overrides/settings.override.json  # settings.json Erweiterungen
```

In `CLAUDE.md` und `AGENTS.md`:

```markdown
<!-- CUSTOM:BEGIN -->
Eigene Anweisungen hier
<!-- CUSTOM:END -->
```

> Vollständige Doku: [docs/layer-system.md](docs/layer-system.md)
```

- [ ] **Schritt 2: Finale Test-Suite ausführen**

```bash
go test ./...
```

Erwartung: alle Tests `PASS`

- [ ] **Schritt 3: Committen**

```bash
git add README.md
git commit -m "feat: add README in forgedeck style"
```

---

## Abschluss

- [ ] **Build-Check**

```bash
go build ./cmd/forgecrate/
./forgecrate --help
```

- [ ] **Alle Tests bestehen**

```bash
go test ./... -v
```

- [ ] **Repo-Struktur validieren**

```bash
find . -type f | grep -v '.git' | sort
```

Erwartung: alle geplanten Dateien vorhanden.
