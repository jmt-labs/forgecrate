# repo-onboarding MCP-Umstellung Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Den `forgecrate-repo-onboarding`-Skill so anpassen, dass er memory-bank explizit via MCP liest und schreibt statt via Read/Write-Datei-Tools.

**Architecture:** Reine Textänderung in zwei SKILL.md-Dateien (Quelle + deployed). Keine Go-Code-Änderungen, keine neuen Dateien.

**Tech Stack:** Markdown, MCP-Tools (`mcp__memory-bank__memory_bank_read`, `mcp__memory-bank__memory_bank_write`, `mcp__memory-bank__memory_bank_update`)

---

### Task 1: base/skills SKILL.md anpassen

**Files:**
- Modify: `base/skills/forgecrate-repo-onboarding/SKILL.md`

- [ ] **Schritt 1: Datei lesen**

```bash
cat base/skills/forgecrate-repo-onboarding/SKILL.md
```

- [ ] **Schritt 2: Zeile ersetzen**

Den letzten Satz in Schritt 5 ersetzen:

Alt:
```
   Schreibe die Dateien mit den Read/Write-Tools direkt.
```

Neu:
```
   Lies jede Zieldatei zuerst via `mcp__memory-bank__memory_bank_read`, um bestehende Inhalte zu kennen. Schreibe neue Inhalte via `mcp__memory-bank__memory_bank_write` (vollständiger Ersatz) oder `mcp__memory-bank__memory_bank_update` (gezielte Ergänzung bestehender Abschnitte). Verwende keine Read/Write-Datei-Tools für memory-bank-Operationen.
```

- [ ] **Schritt 3: Diff prüfen**

```bash
git diff base/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartung: Nur die eine Zeile geändert, nichts anderes.

- [ ] **Schritt 4: Commit**

```bash
git add base/skills/forgecrate-repo-onboarding/SKILL.md
git commit -m "feat(skill): repo-onboarding liest und schreibt memory-bank via MCP"
```

---

### Task 2: deployed .claude/skills SKILL.md synchronisieren

**Files:**
- Modify: `.claude/skills/forgecrate-repo-onboarding/SKILL.md`

- [ ] **Schritt 1: Dieselbe Ersetzung wie in Task 1 durchführen**

Den letzten Satz in Schritt 5 identisch ersetzen:

Alt:
```
   Schreibe die Dateien mit den Read/Write-Tools direkt.
```

Neu:
```
   Lies jede Zieldatei zuerst via `mcp__memory-bank__memory_bank_read`, um bestehende Inhalte zu kennen. Schreibe neue Inhalte via `mcp__memory-bank__memory_bank_write` (vollständiger Ersatz) oder `mcp__memory-bank__memory_bank_update` (gezielte Ergänzung bestehender Abschnitte). Verwende keine Read/Write-Datei-Tools für memory-bank-Operationen.
```

- [ ] **Schritt 2: Diff prüfen**

```bash
git diff .claude/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartung: Inhalt identisch mit `base/skills/...`.

- [ ] **Schritt 3: Commit**

```bash
git add .claude/skills/forgecrate-repo-onboarding/SKILL.md
git commit -m "chore: deployed skill mit base synchronisieren"
```

---

### Task 3: Verifikation

- [ ] **Schritt 1: Beide Dateien inhaltlich vergleichen**

```bash
diff base/skills/forgecrate-repo-onboarding/SKILL.md \
     .claude/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartung: Kein Output (identisch).

- [ ] **Schritt 2: Alte Tool-Referenz ist weg**

```bash
grep -n "Read/Write-Tools" \
  base/skills/forgecrate-repo-onboarding/SKILL.md \
  .claude/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartung: Kein Treffer.

- [ ] **Schritt 3: Neue MCP-Tool-Referenzen vorhanden**

```bash
grep -n "memory_bank_read\|memory_bank_write\|memory_bank_update" \
  base/skills/forgecrate-repo-onboarding/SKILL.md \
  .claude/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartung: Jeweils 3 Treffer (ein Tool pro Zeile, in beiden Dateien).
