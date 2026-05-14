# Base Layer: MCP Server — Design

**Datum:** 2026-05-14
**Issue:** #3
**Branch:** `feat/extensions`

## Ziel

Vier universell nützliche MCP Server werden im base layer deklariert. Sie stehen in jedem Repo, das claude-setup verwendet, automatisch zur Verfügung. Alle Server sind kostenlos, kein externer Bezahldienst.

## Abhängigkeit

Die Deklaration erfolgt in `base/extensions.yaml`. Das Extensions-Mechanismus-System (Issue #1) muss `transport: http` für den GitHub MCP unterstützen — das erfordert eine Erweiterung des bestehenden `MCP`-Typs in `internal/extensions/extensions.go`.

---

## Server

### 1. GitHub MCP (`github/github-mcp-server`)

**Konfiguration:**
```yaml
mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/
    scope: local
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: "${GITHUB_PERSONAL_ACCESS_TOKEN}"
```

`claude mcp add --transport http github https://api.githubcopilot.com/mcp/`

**Wann Claude es verwendet:**
- Für alle Operationen mit GitHub: Issues lesen, erstellen, kommentieren; PRs öffnen, reviewen, mergen; Code über Repos hinweg suchen; Commits und Branches prüfen
- Für Workflow-Automatisierung: Labels setzen, Milestones zuweisen, Checks prüfen

**Wann Claude es NICHT verwendet:**
- Lokale Dateioperationen — dafür direkte Tools (Read, Edit, Bash)
- Lokale Git-Operationen — dafür Bash mit git-Befehlen

**Voraussetzung:** `GITHUB_PERSONAL_ACCESS_TOKEN` als Umgebungsvariable (kostenloser GitHub-Account reicht).

---

### 2. Fetch (`@modelcontextprotocol/server-fetch`)

**Konfiguration:**
```yaml
mcp:
  - name: fetch
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-fetch"]
```

**Wann Claude es verwendet:**
- Externe Dokumentation abrufen: Library-Docs, MDN, RFC-Spezifikationen, API-Referenzen
- Webinhalte für Recherche: Blogposts, Changelogs, Release Notes
- Inhalte von URLs aus Issues oder PRs nachlesen

**Wann Claude es NICHT verwendet:**
- GitHub-Inhalte — dafür GitHub MCP (strukturierter Zugriff)
- Lokale Dateien — dafür direkte Tools

**Voraussetzung:** Node.js / npx (standard in Entwicklungsumgebungen).

---

### 3. Memory (`@modelcontextprotocol/server-memory`)

**Konfiguration:**
```yaml
mcp:
  - name: memory
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-memory"]
    env:
      MEMORY_FILE_PATH: ".claude/memory.json"
```

Die Datei `.claude/memory.json` liegt im Repo und wird mit Git versioniert — kein sensitives Material speichern.

**Wann Claude es verwendet:**
- **Schreiben** — nach wichtigen Entscheidungen: Architekturwahl, Begründung für nicht-offensichtliche Lösungen, Ergebnis eines Debugging-Falls; nach Brainstorming-Sessions
- **Lesen** — zu Beginn einer neuen Session oder nach Context-Kompaktierung; wenn unklar ist, warum etwas so gebaut wurde

**Struktur der Einträge:**
- Entitäten: Komponenten, Module, Features
- Observationen: Entscheidungen mit Begründung, bekannte Einschränkungen, offene Fragen
- Relationen: Abhängigkeiten zwischen Komponenten

**Was NICHT gespeichert wird:**
- API-Keys, Tokens, Passwörter
- Temporärer Zwischenstand (gehört in Tasks, nicht Memory)
- Dinge die aus dem Code direkt lesbar sind

---

### 4. context-mode (`mksglu/context-mode`)

**Konfiguration:**
```yaml
mcp:
  - name: context-mode
    scope: local
    command: npx
    args: ["-y", "context-mode"]
```

**Wann Claude es verwendet:**
- **Automatisch aktiv** für alle Tool-Aufrufe — kein expliziter Aufruf nötig; der Server sandboxt Tool-Output selbständig
- `ctx_search` — bei kompaktierten Sessions: relevanten Kontext aus der History wiederfinden (BM25-Suche über alle vergangenen Tool-Outputs)
- `ctx_insight` — Überblick über bisherigen Session-Verlauf und Entscheidungen
- `ctx_stats` — prüfen wie viel Context gespart wurde
- `ctx_doctor` — bei Problemen mit dem Server selbst

**Wann Claude es NICHT explizit aufruft:**
- Für normale Datei- oder Bash-Operationen — diese werden automatisch sandboxt, kein manueller Aufruf nötig

**Lizenz:** Elastic License 2.0 (source-available, kostenlos nutzbar).

---

## Erweiterung des Extensions-Mechanismus

Der bestehende `MCP`-Typ in `internal/extensions/extensions.go` unterstützt nur `command`/`args` (stdio). Für GitHub MCP wird `transport: http` und `url` benötigt.

Erweiterter Typ:
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

Install-Befehl je nach Transport:
- stdio: `claude mcp add <name> --scope <scope> <command> [args...]`
- http: `claude mcp add --transport http <name> <url> --scope <scope>`

---

## Datei-Änderungen

| Datei | Änderung |
|---|---|
| `base/extensions.yaml` | 4 MCP Server hinzufügen |
| `internal/extensions/extensions.go` | `transport` und `url` Felder ergänzen |
| `internal/extensions/install.go` | HTTP-Transport-Zweig in `Install()` |
| `base/CLAUDE.md` | Abschnitt "MCP Server Verwendung" im GENERATED-Block |
