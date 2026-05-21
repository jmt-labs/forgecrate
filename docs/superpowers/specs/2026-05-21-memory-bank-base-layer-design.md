# Design: memory-bank im base layer

**Datum:** 2026-05-21  
**Status:** Approved

## Ziel

Der `memory-bank` MCP-Server wird verpflichtender Bestandteil des base layers. Jedes Repo, das mit `forgecrate init` oder `forgecrate update` aufgesetzt wird, erhält automatisch:

1. Den MCP-Server-Eintrag in `.mcp.json`
2. Das `memory-bank/`-Verzeichnis mit Starter-Dateien
3. Dokumentation in `CLAUDE.md`

Der `forgecrate-repo-onboarding`-Skill befüllt die Starter-Dateien nach der Repo-Analyse automatisch (ohne Rückfrage).

## Komponenten

### 1. `base/extensions.yaml` — MCP-Eintrag

```yaml
- name: memory-bank
  command: npx
  args: ["-y", "@allpepper/memory-bank-mcp"]
  env:
    MEMORY_BANK_ROOT: "memory-bank"
```

Wird wie alle anderen MCP-Einträge von `installExtensions()` in `deploy.go` zu `.mcp.json` generiert.

### 2. `base/memory-bank/` — Template-Dateien

Fünf Markdown-Dateien mit deutschen Platzhaltern (Kommentare im HTML-Kommentar-Format):

| Datei | Zweck |
|---|---|
| `projectbrief.md` | Was ist das Projekt, Ziele, Nicht-Ziele |
| `techContext.md` | Stack, Tools, Constraints |
| `systemPatterns.md` | ADRs, Muster, Anti-Patterns |
| `activeContext.md` | Aktueller Fokus, offene Fragen, Blocker |
| `progress.md` | Fertig, in Arbeit, nächste Schritte |

### 3. `deploy.go` — `scaffoldMemoryBank()`

Neuer Schritt in `RunWithClaude()` nach `installExtensions()`:

- Quelle: `<sourceDir>/base/memory-bank/`
- Ziel: `<destDir>/memory-bank/`
- **Idempotent:** Vorhandene Dateien werden nicht überschrieben — nur fehlende Dateien werden angelegt.
- Kein Hash-Tracking (kein managed file).

### 4. `base/CLAUDE.md` — Dokumentationsabschnitt

Neuer Abschnitt `### Memory-Bank (\`memory-bank\`)`:

```markdown
### Memory-Bank (`memory-bank`)

Strukturiertes, dateibasiertes Projektgedächtnis in `memory-bank/`. Persistiert
kontextuelles Wissen über Sessions hinweg.

**Schreiben nach:** Architekturentscheidungen, Projektstruktur-Überblick, aktueller
Fortschritt, offene Fragen.

**Lesen am:** Sessionbeginn, nach Context-Kompaktierung.

**Dateien:**
- `projectbrief.md` — Projektziel und Scope
- `techContext.md` — Stack, Tools, Constraints
- `systemPatterns.md` — ADRs und Muster
- `activeContext.md` — Aktueller Fokus und Blocker
- `progress.md` — Fortschritt und nächste Schritte
```

Außerdem: Zahl „Vier MCP-Server" → „Fünf MCP-Server" im Einleitungssatz.

### 5. `base/skills/forgecrate-repo-onboarding/skill.md` — memory-bank befüllen

Neuer Abschnitt nach dem CLAUDE.md-Schritt:

**Ablauf:**
1. Aus der bereits durchgeführten Repo-Analyse `projectbrief.md` befüllen (Projektbeschreibung, Ziele, Nicht-Ziele)
2. `techContext.md` befüllen (erkannter Stack, Build/Test/Lint-Tools, Infrastruktur)
3. `systemPatterns.md` befüllen (erkannte Architektur-Muster, wichtige ADRs aus git log / README)
4. `activeContext.md` und `progress.md` — Template bleibt leer (nicht ableitbar aus statischer Analyse)

**Verhalten:** Direkt schreiben, keine Bestätigungsfrage.  
**Idempotent:** Dateien werden überschrieben (der Skill läuft nur einmalig nach `forgecrate init`).

## Datenfluss

```
forgecrate init
  └── deploy.Run()
        ├── installExtensions()   → .mcp.json (inkl. memory-bank)
        ├── scaffoldMemoryBank()  → memory-bank/*.md (nur neue Dateien)
        └── copySkills()          → .claude/skills/forgecrate-repo-onboarding/

Nutzer ruft forgecrate-repo-onboarding auf
  ├── Repo analysieren
  ├── CLAUDE.md-Vorschlag → Nutzer bestätigt → GENERATED-Block ersetzen
  └── memory-bank befüllen (direkt, ohne Rückfrage)
        ├── projectbrief.md  ← Analyse-Ergebnis
        ├── techContext.md   ← Analyse-Ergebnis
        ├── systemPatterns.md ← Analyse-Ergebnis
        ├── activeContext.md  ← leer (Template)
        └── progress.md       ← leer (Template)
```

## Abgrenzung

- `memory` (MCP) — Graph-basiertes, projektübergreifendes Gedächtnis (`.claude/memory.json`)
- `memory-bank` (MCP) — Dateibasiertes, repo-spezifisches Gedächtnis (`memory-bank/*.md`)

Beide bleiben unabhängig und ergänzen sich.

## Tests

- Unit-Test für `scaffoldMemoryBank()`: bestehende Dateien werden nicht überschrieben
- Unit-Test: fehlende Dateien werden angelegt
- E2E: `forgecrate init` → `memory-bank/` existiert mit 5 Dateien
