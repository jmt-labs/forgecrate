# Base Layer: MCP Server — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Vier MCP-Server im base layer deklarieren und den Extensions-Mechanismus um HTTP-Transport erweitern, damit GitHub MCP via `transport: http` installiert wird.

**Architecture:** `internal/extensions/extensions.go` erhält `Transport` und `URL` Felder im `MCP`-Typ. `install.go` bekommt einen HTTP-Zweig (`claude mcp add --transport http <name> <url> --scope <scope>`). `base/extensions.yaml` deklariert alle vier Server. `base/CLAUDE.md` erhält einen neuen Abschnitt über MCP-Verwendung im GENERATED-Block.

**Tech Stack:** Go 1.24, `gopkg.in/yaml.v3`, `os/exec`

**Abhängigkeit:** Dieser Plan setzt voraus, dass `docs/superpowers/plans/2026-05-14-extensions.md` (Tasks 1–4) bereits implementiert ist und `internal/extensions/` existiert.

---

## Dateistruktur

| Datei | Aktion | Zweck |
|---|---|---|
| `internal/extensions/extensions.go` | Ändern | `Transport` und `URL` Felder ergänzen |
| `internal/extensions/extensions_test.go` | Ändern | Tests für HTTP-Transport-Parsing |
| `internal/extensions/install.go` | Ändern | HTTP-Transport-Zweig in `Install()` |
| `internal/extensions/install_test.go` | Ändern | Test für HTTP-MCP-Aufruf |
| `base/extensions.yaml` | Ändern | 4 MCP-Server hinzufügen |
| `base/CLAUDE.md` | Ändern | Abschnitt "MCP Server" im GENERATED-Block |

---

### Task 1: `MCP`-Typ um `Transport` und `URL` erweitern

**Files:**
- Modify: `internal/extensions/extensions.go`
- Modify: `internal/extensions/extensions_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

Füge in `internal/extensions/extensions_test.go` nach `TestLoad` ein:

```go
func TestLoadHTTPTransport(t *testing.T) {
	dir := t.TempDir()
	content := `
mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/
    scope: local
`
	os.WriteFile(filepath.Join(dir, "extensions.yaml"), []byte(content), 0644)

	ext, err := extensions.Load(filepath.Join(dir, "extensions.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ext.MCP) != 1 {
		t.Fatalf("expected 1 MCP, got %d", len(ext.MCP))
	}
	if ext.MCP[0].Transport != "http" {
		t.Errorf("Transport: got %q, want %q", ext.MCP[0].Transport, "http")
	}
	if ext.MCP[0].URL != "https://api.githubcopilot.com/mcp/" {
		t.Errorf("URL: got %q", ext.MCP[0].URL)
	}
}
```

- [ ] **Schritt 2: Tests ausführen — müssen FAIL**

```
go test ./internal/extensions/... -run TestLoadHTTPTransport -v
```

Erwartet: `FAIL` mit `ext.MCP[0].Transport: got "", want "http"` (Feld existiert nicht).

- [ ] **Schritt 3: `MCP`-Typ in `extensions.go` erweitern**

Ersetze den `MCP`-Struct:

```go
type MCP struct {
	Name      string            `yaml:"name"`
	Scope     string            `yaml:"scope"`
	Transport string            `yaml:"transport"` // "stdio" (default) oder "http"
	URL       string            `yaml:"url"`        // nur für transport: http
	Command   string            `yaml:"command"`    // nur für transport: stdio
	Args      []string          `yaml:"args"`
	Env       map[string]string `yaml:"env"`
}
```

- [ ] **Schritt 4: Tests ausführen — müssen PASS**

```
go test ./internal/extensions/... -v
```

Erwartet: alle Tests `PASS`.

- [ ] **Schritt 5: Commit**

```bash
git add internal/extensions/extensions.go internal/extensions/extensions_test.go
git commit -m "feat(extensions): add Transport and URL fields to MCP type"
```

---

### Task 2: HTTP-Transport-Zweig im Installer

**Files:**
- Modify: `internal/extensions/install.go`
- Modify: `internal/extensions/install_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

Füge in `internal/extensions/install_test.go` hinzu:

```go
func TestInstallerMCPHTTP(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{
				Name:      "github",
				Transport: "http",
				URL:       "https://api.githubcopilot.com/mcp/",
				Scope:     "local",
			},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	got := string(data)
	if !strings.Contains(got, "mcp add --transport http github https://api.githubcopilot.com/mcp/ --scope local") {
		t.Errorf("expected http mcp add call, got: %q", got)
	}
}

func TestInstallerMCPHTTPEnv(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		MCP: []extensions.MCP{
			{
				Name:      "github",
				Transport: "http",
				URL:       "https://api.githubcopilot.com/mcp/",
				Scope:     "local",
				Env:       map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": "tok"},
			},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Env wird übergeben — kein direkter Weg es über argsFile zu prüfen,
	// daher nur sicherstellen dass der Aufruf ohne Fehler durchläuft.
	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "mcp add --transport http") {
		t.Errorf("expected http transport call, got: %q", string(data))
	}
}
```

- [ ] **Schritt 2: Tests ausführen — müssen FAIL**

```
go test ./internal/extensions/... -run TestInstallerMCPHTTP -v
```

Erwartet: `FAIL` — aktuell wird kein `--transport http` erzeugt.

- [ ] **Schritt 3: HTTP-Zweig in `install.go` einbauen**

Ersetze die MCP-Schleife in `Install()`:

```go
for _, m := range ext.MCP {
	scope := m.Scope
	if scope == "" {
		scope = "local"
	}

	var args []string
	if m.Transport == "http" {
		args = []string{"mcp", "add", "--transport", "http", m.Name, m.URL, "--scope", scope}
	} else {
		args = []string{"mcp", "add", m.Name, "--scope", scope, m.Command}
		args = append(args, m.Args...)
	}

	cmd := exec.Command(claude, args...)
	cmd.Env = append(os.Environ(), envPairs(m.Env)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("warn: mcp add %s: %v: %s", m.Name, err, out)
	}
}
```

- [ ] **Schritt 4: Tests ausführen — müssen PASS**

```
go test ./internal/extensions/... -v
```

Erwartet: alle Tests `PASS`.

- [ ] **Schritt 5: Gesamte Test-Suite**

```
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 6: Commit**

```bash
git add internal/extensions/install.go internal/extensions/install_test.go
git commit -m "feat(extensions): support HTTP transport in Installer"
```

---

### Task 3: `base/extensions.yaml` mit 4 MCP-Servern befüllen

**Files:**
- Modify: `base/extensions.yaml`

- [ ] **Schritt 1: Datei schreiben**

Vollständiger Inhalt von `base/extensions.yaml`:

```yaml
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers

mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/
    scope: local

  - name: fetch
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-fetch"]

  - name: memory
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-memory"]
    env:
      MEMORY_FILE_PATH: ".claude/memory.json"

  - name: context-mode
    scope: local
    command: npx
    args: ["-y", "context-mode"]
```

- [ ] **Schritt 2: YAML-Parsing manuell prüfen**

```bash
go run -e . 2>/dev/null || true
go test ./internal/extensions/... -v -run TestLoad
```

Erwartet: Tests laufen durch (Load-Logik ist format-agnostisch).

- [ ] **Schritt 3: Commit**

```bash
git add base/extensions.yaml
git commit -m "feat(base): declare 4 MCP servers in extensions.yaml"
```

---

### Task 4: `base/CLAUDE.md` — MCP-Verwendung dokumentieren

**Files:**
- Modify: `base/CLAUDE.md`

- [ ] **Schritt 1: Aktuellen Inhalt lesen**

```
cat base/CLAUDE.md
```

Suche den Block `<!-- GENERATED:END -->`. Der neue Abschnitt wird direkt davor eingefügt.

- [ ] **Schritt 2: Abschnitt einfügen**

Im `<!-- GENERATED:BEGIN -->` Block, direkt vor `<!-- GENERATED:END -->`, einfügen:

```markdown
## MCP Server

Vier MCP-Server sind im base layer deklariert und stehen automatisch zur Verfügung.

### GitHub (`github`)

Für alle Operationen mit GitHub: Issues, PRs, Code-Suche, Branches, Checks, Labels.

**Verwende es für:** Issues lesen/erstellen/kommentieren, PRs öffnen/reviewen/mergen, Code repo-übergreifend suchen, Workflow-Labels setzen.

**Verwende es NICHT für:** Lokale Dateioperationen (→ Read/Edit/Bash), lokale Git-Kommandos (→ Bash mit git).

**Voraussetzung:** `GITHUB_PERSONAL_ACCESS_TOKEN` als Umgebungsvariable.

### Fetch (`fetch`)

Externe Webinhalte abrufen: Dokumentation, MDN, RFCs, Changelogs, Release Notes, URLs aus Issues.

**Verwende es NICHT für:** GitHub-Inhalte (→ github MCP), lokale Dateien (→ Read).

### Memory (`memory`)

Projektübergreifendes Wissen persistent speichern. Datei: `.claude/memory.json` (versioniert).

**Schreiben nach:** Architekturentscheidungen, Begründungen für nicht-offensichtliche Lösungen, Debugging-Ergebnisse, Brainstorming-Ergebnisse.

**Lesen am:** Sessionbeginn, nach Context-Kompaktierung, wenn unklar warum etwas so gebaut wurde.

**Niemals speichern:** API-Keys, Tokens, Passwörter, temporärer Zwischenstand, Code-Details die direkt aus dem Code lesbar sind.

### Context Mode (`context-mode`)

Sandboxt Tool-Output automatisch — kein expliziter Aufruf nötig.

**Explizit aufrufen:**
- `ctx_search` — nach Context-Kompaktierung: relevante Infos aus der Session-History finden (BM25-Suche)
- `ctx_insight` — Überblick über bisherigen Session-Verlauf
- `ctx_stats` — gespartes Context-Budget prüfen
- `ctx_doctor` — bei Problemen mit dem Server
```

- [ ] **Schritt 3: Gesamte Test-Suite**

```
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 4: Commit**

```bash
git add base/CLAUDE.md
git commit -m "feat(base): document MCP server usage in CLAUDE.md"
```
