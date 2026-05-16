# TDD Bug-Regression & Agenten-Identität Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Zwei CLAUDE.md-Ergänzungen: Regressionspflicht für Bugfixes im TDD-Flavor und Naming/Color-Regeln für Subagenten.

**Architecture:** Reine Textänderungen in zwei Konfigurationsdateien (`flavors/tdd/CLAUDE.md` und `base/CLAUDE.md`). Keine Code-, Skill- oder Teständerungen nötig.

**Tech Stack:** Markdown, Git

---

### Task 1: Regressions-Regel im TDD-Flavor ergänzen

**Files:**
- Modify: `flavors/tdd/CLAUDE.md`

- [ ] **Step 1: Neue Zeile in flavors/tdd/CLAUDE.md einfügen**

Datei aktuell:
```markdown
## TDD-Flavor

- Test schreiben → ausführen (muss fehlschlagen) → implementieren → ausführen (muss bestehen) → committen
- Kein Produktionscode ohne vorherigen Test
- Test-Namen beschreiben Verhalten, nicht Implementierung
- Mocks nur an Systemgrenzen (externe APIs, Datenbanken)
```

Nach der Änderung:
```markdown
## TDD-Flavor

- Test schreiben → ausführen (muss fehlschlagen) → implementieren → ausführen (muss bestehen) → committen
- Kein Produktionscode ohne vorherigen Test
- Test-Namen beschreiben Verhalten, nicht Implementierung
- Mocks nur an Systemgrenzen (externe APIs, Datenbanken)
- Jeder gefundene Bug erhält vor dem Fix einen Regressionstest
```

- [ ] **Step 2: Änderung prüfen**

```bash
cat flavors/tdd/CLAUDE.md
```

Erwartete Ausgabe: Datei endet mit der Zeile `- Jeder gefundene Bug erhält vor dem Fix einen Regressionstest`.

- [ ] **Step 3: Committen**

```bash
git add flavors/tdd/CLAUDE.md
git commit -m "feat(tdd): Regressionspflicht für Bugfixes ergänzen"
```

---

### Task 2: Pflicht-Skills-Tabelle in base/CLAUDE.md erweitern

**Files:**
- Modify: `base/CLAUDE.md` (Zeile 14 — nach der Debug-Zeile)

- [ ] **Step 1: Neue Zeile in Pflicht-Skills-Tabelle einfügen**

Tabelle aktuell (Zeilen 9–15):
```markdown
| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
```

Nach der Änderung:
```markdown
| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |
```

- [ ] **Step 2: Änderung prüfen**

```bash
grep -A6 "## Pflicht-Skills" base/CLAUDE.md
```

Erwartete Ausgabe: Tabelle enthält die neue Zeile mit `Bug gefunden (nach Debug)`.

- [ ] **Step 3: Committen**

```bash
git add base/CLAUDE.md
git commit -m "feat(base): Pflicht-Skill für Bug-Regressionstests ergänzen"
```

---

### Task 3: Agenten-Identität in base/CLAUDE.md ergänzen

**Files:**
- Modify: `base/CLAUDE.md` (nach Zeile 60 — Ende der Parallelisierungs-Tabelle)

- [ ] **Step 1: Neue Untersektion nach "Im Zweifelsfall Background nutzen..." einfügen**

Abschnitt aktuell (Ende von "Parallelisierung & Isolation"):
```markdown
Im Zweifelsfall Background nutzen — warten ist kein Default.

## MCP Server
```

Nach der Änderung:
```markdown
Im Zweifelsfall Background nutzen — warten ist kein Default.

### Agenten-Identität

Jeder Subagent bekommt:
- **Eindeutigen Namen** — via `description`-Parameter im Agent-Tool-Aufruf (3–5 Wörter, Rolle + Aufgabe)
- **Eindeutige Farbe** — dynamisch durch FleetView zugewiesen; keine zwei gleichzeitig laufenden Agenten teilen eine Farbe

## MCP Server
```

- [ ] **Step 2: Änderung prüfen**

```bash
grep -A6 "Agenten-Identität" base/CLAUDE.md
```

Erwartete Ausgabe: Neue Untersektion mit den beiden Bullet-Points.

- [ ] **Step 3: Committen**

```bash
git add base/CLAUDE.md
git commit -m "feat(base): Agenten-Identität mit eindeutigem Namen und Farbe ergänzen"
```

---

### Task 4: Deployment testen

**Files:** keine Änderungen

- [ ] **Step 1: Build ausführen**

```bash
go build ./...
```

Erwartete Ausgabe: kein Output (erfolgreich).

- [ ] **Step 2: Tests ausführen**

```bash
go test ./...
```

Erwartete Ausgabe: `ok  github.com/jmt-labs/claude-setup/cmd/claude-setup`

- [ ] **Step 3: Finaler Commit falls nötig**

Nur falls Step 1/2 Probleme aufgedeckt haben. Sonst: Plan abgeschlossen.
