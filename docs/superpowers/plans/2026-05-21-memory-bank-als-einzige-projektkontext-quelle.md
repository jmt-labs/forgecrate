# memory-bank als einzige Projektkontext-Quelle — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Projektspezifischen Inline-Inhalt aus dem GENERATED-Block von CLAUDE.md entfernen; repo-onboarding schreibt nur noch in memory-bank; handoff schreibt via memory-bank MCP statt HANDOFF.md.

**Architecture:** Drei Markdown-Dateien werden editiert: `base/CLAUDE.md`, `base/skills/forgecrate-repo-onboarding/SKILL.md`, `base/skills/forgecrate-handoff/SKILL.md`. Kein Go-Code ändert sich. Anschließend wird dieses Repo selbst migriert (CLAUDE.md-GENERATED-Block bereinigen, memory-bank befüllen).

**Tech Stack:** Markdown-Dateien, memory-bank MCP-Tools (`memory_bank_write`, `memory_bank_update`), git

---

### Task 1: `base/CLAUDE.md` — Session-Start und Projektkontext-Abschnitt

**Files:**
- Modify: `base/CLAUDE.md` (Abschnitt `## Session-Start` + neuer Abschnitt `## Projektkontext`)

- [ ] **Schritt 1: `## Session-Start` aktualisieren**

  Aktuell steht dort:
  ```
  ## Session-Start

  Beim Session-Start: `ls HANDOFF.md 2>/dev/null` ausführen. Falls vorhanden: Datei lesen und als Kontext verwenden, dann fragen: „HANDOFF.md gefunden und gelesen. Soll ich sie löschen?"
  ```

  Ersetzen durch:
  ```
  ## Session-Start

  Beim Session-Start: memory-bank via MCP lesen um den aktuellen Projektkontext zu verstehen.
  ```

- [ ] **Schritt 2: Neuen Abschnitt `## Projektkontext` am Ende von `base/CLAUDE.md` anhängen**

  Am Ende der Datei (nach dem letzten Abschnitt) einfügen:
  ```markdown
  ## Projektkontext

  Nutze den `memory-bank` MCP-Server um den aktuellen Projektkontext zu lesen.
  ```

- [ ] **Schritt 3: Änderung prüfen**

  ```bash
  grep -A3 "Session-Start\|Projektkontext" base/CLAUDE.md
  ```
  Erwartet: Kein Verweis auf `HANDOFF.md`; `## Projektkontext` mit MCP-Verweis vorhanden.

- [ ] **Schritt 4: Committen**

  ```bash
  git add base/CLAUDE.md
  git commit -m "feat(base): session-start auf memory-bank umstellen, projektkontext-verweis ergänzen"
  ```

---

### Task 2: `base/skills/forgecrate-repo-onboarding/SKILL.md` — CLAUDE.md-Schritte entfernen

**Files:**
- Modify: `base/skills/forgecrate-repo-onboarding/SKILL.md`

Aktueller Aufbau des Skills:
- Schritte 1–4: Repo analysieren ✓ bleibt
- Schritt 5: CLAUDE.md-Vorschlag erstellen ✗ entfällt
- Schritt 6: Nutzer fragen ob CLAUDE.md ersetzt werden soll ✗ entfällt
- Schritt 7: memory-bank befüllen → wird zu Schritt 5 umnummeriert

- [ ] **Schritt 1: Schritte 5 und 6 aus dem Skill entfernen**

  Den kompletten Block von `5. **CLAUDE.md-Vorschlag erstellen**` bis zum Ende von Schritt 6 löschen (inkl. des Markdown-Codeblocks mit dem Template und der Übergabe-Frage).

  Das ist der zu entfernende Block:
  ```
  5. **CLAUDE.md-Vorschlag erstellen** — erzeuge Text für den `<!-- GENERATED:BEGIN -->…<!-- GENERATED:END -->`-Block:

  ```markdown
  ## Projekt-Übersicht
  ...
  ## Externe Abhängigkeiten
  ...
  ```

  6. **Übergabe** — zeige den Vorschlag und frage: "Soll ich den GENERATED-Block in `CLAUDE.md` damit ersetzen?"
  ```

- [ ] **Schritt 2: Alten Schritt 7 zu Schritt 5 umnummerieren**

  `7. **memory-bank befüllen**` → `5. **memory-bank befüllen**`

- [ ] **Schritt 3: Abschlusstext des Skills aktualisieren**

  Letzten Satz von Schritt 5 (ehemals 7) anpassen. Aktuell endet er mit:
  ```
  Schreibe die Dateien mit den Read/Write-Tools direkt. Kein Prompt an den Nutzer.
  ```

  Ersetzen durch:
  ```
  Schreibe die Dateien mit den Read/Write-Tools direkt.
  Anschließend: kurze Bestätigung ausgeben welche memory-bank-Dateien befüllt wurden (ein Satz je Datei).
  ```

- [ ] **Schritt 4: Skill inhaltlich prüfen**

  ```bash
  cat base/skills/forgecrate-repo-onboarding/SKILL.md
  ```
  Erwartet: Kein Verweis mehr auf `CLAUDE.md`, nur noch 5 Schritte, letzter Schritt ist memory-bank.

- [ ] **Schritt 5: Committen**

  ```bash
  git add base/skills/forgecrate-repo-onboarding/SKILL.md
  git commit -m "feat(skill): repo-onboarding schreibt nur noch in memory-bank"
  ```

---

### Task 3: `base/skills/forgecrate-handoff/SKILL.md` — Vollständig auf memory-bank MCP umschreiben

**Files:**
- Modify: `base/skills/forgecrate-handoff/SKILL.md`

- [ ] **Schritt 1: Skill vollständig ersetzen**

  Den gesamten Inhalt von `base/skills/forgecrate-handoff/SKILL.md` durch folgenden Inhalt ersetzen:

  ````markdown
  # Handoff

  Aktualisiert die memory-bank mit dem aktuellen Session-Kontext für AI-Modellwechsel oder Session-Übergabe. Kein externes Tool nötig, kein HANDOFF.md.

  ## Ablauf

  **Schritt 1 — Daten sammeln (parallel ausführen):**

  ```bash
  # Git-Info
  git branch --show-current && git log --oneline -10 && git status --short
  date "+%Y-%m-%d %H:%M:%S"

  # TODOs und FIXMEs
  grep -rEn '\b(TODO|FIXME)[:([ ]' \
    --include="*.go" --include="*.ts" --include="*.tsx" \
    --include="*.py" --include="*.rs" --include="*.js" \
    . 2>/dev/null | grep -v node_modules | grep -v ".git" | head -20
  ```

  **Schritt 2 — `activeContext.md` via memory-bank MCP schreiben:**

  Tool: `memory_bank_write` mit `file_name: "activeContext.md"` und folgendem Inhalt:

  ```
  # Active Context

  ## Aktueller Branch
  <Branch-Name>

  ## Uncommitted Changes
  <git status --short Output, oder "Working tree clean">

  ## Offene Fragen / Blocker
  <Leer lassen — wird manuell gepflegt>
  ```

  **Schritt 3 — `progress.md` via memory-bank MCP schreiben:**

  Tool: `memory_bank_write` mit `file_name: "progress.md"` und folgendem Inhalt:

  ```
  # Progress

  ## Recent Activity
  <letzte 5–10 Commits: `<hash>` <message>>

  ## Known Issues
  <TODO/FIXME mit file:line — Abschnitt weglassen wenn keine gefunden>

  ## Was als nächstes kommt
  <Leer lassen — wird manuell gepflegt>
  ```

  **Schritt 4 — Abschluss:**

  Dem Nutzer bestätigen: welche Dateien wurden in der memory-bank aktualisiert (ein Satz je Datei).
  ````

- [ ] **Schritt 2: Datei prüfen**

  ```bash
  cat base/skills/forgecrate-handoff/SKILL.md
  ```
  Erwartet: Kein Verweis auf `HANDOFF.md`, `memory_bank_write` als MCP-Tool, zwei Zieldateien (`activeContext.md`, `progress.md`).

- [ ] **Schritt 3: Committen**

  ```bash
  git add base/skills/forgecrate-handoff/SKILL.md
  git commit -m "feat(skill): handoff schreibt via memory-bank mcp statt HANDOFF.md"
  ```

---

### Task 4: Dieses Repo migrieren

Die lokale `CLAUDE.md` (forgecrate selbst) hat `## Projekt-Übersicht`, `## Struktur`, `## Workflow`, `## Externe Abhängigkeiten` im GENERATED-Block. Diese werden durch `forgecrate update` überschrieben. Vorher in memory-bank überführen.

**Files:**
- Modify: `memory-bank/techContext.md`, `memory-bank/systemPatterns.md`
- Modify: `CLAUDE.md` (GENERATED-Block bereinigen via forgecrate update oder manuell)

- [ ] **Schritt 1: Bestehenden Inhalt aus CLAUDE.md in memory-bank-Dateien schreiben**

  Tool: `memory_bank_update` mit `file_name: "techContext.md"` — den aktuellen Stack-Inhalt aus `## Projekt-Übersicht` und `## Workflow` eintragen.

  Tool: `memory_bank_update` mit `file_name: "systemPatterns.md"` — den Inhalt aus `## Struktur` und `## Externe Abhängigkeiten` eintragen.

- [ ] **Schritt 2: GENERATED-Block in lokaler CLAUDE.md bereinigen**

  Die Abschnitte `## Projekt-Übersicht`, `## Struktur`, `## Workflow`, `## Externe Abhängigkeiten` aus dem GENERATED-Block der lokalen `CLAUDE.md` entfernen (manuell via Edit, da `forgecrate update` derzeit noch die alte Logik hat).

  Der GENERATED-Block endet nach dem letzten MCP-Abschnitt (`## MCP-Konfiguration: Single Source of Truth`). Die vier Projekt-Abschnitte danach löschen.

  Dann den neuen `## Projektkontext`-Abschnitt am Ende des GENERATED-Blocks einfügen:
  ```markdown
  ## Projektkontext

  Nutze den `memory-bank` MCP-Server um den aktuellen Projektkontext zu lesen.
  ```

- [ ] **Schritt 3: CUSTOM-Block prüfen**

  ```bash
  grep -n "CUSTOM\|Architektur\|Entwicklungskommandos" CLAUDE.md
  ```
  Der CUSTOM-Block enthält noch "Entwicklungskommandos" und "Architektur" — diese bleiben dort als manuelle Einträge (sie sind projektspezifisch und wurden manuell gepflegt).

- [ ] **Schritt 4: Committen**

  ```bash
  git add CLAUDE.md memory-bank/techContext.md memory-bank/systemPatterns.md
  git commit -m "chore: projektkontext aus CLAUDE.md in memory-bank migrieren"
  ```

---

### Task 5: Verifikation

- [ ] **Schritt 1: Alle geänderten Dateien prüfen**

  ```bash
  grep -n "HANDOFF\|Projekt-Übersicht\|Struktur\|Workflow\|Externe Abhängigkeiten" \
    base/CLAUDE.md \
    base/skills/forgecrate-repo-onboarding/SKILL.md \
    base/skills/forgecrate-handoff/SKILL.md \
    CLAUDE.md
  ```
  Erwartet: Kein Output (keine dieser Begriffe in den geänderten Dateien).

- [ ] **Schritt 2: Go-Tests ausführen (Smoke-Check)**

  ```bash
  go test ./...
  ```
  Erwartet: Alle Tests grün. (Die Go-Tests prüfen die Compose-Logik, kein Skill-Inhalt — aber sicherstellen dass nichts gebrochen wurde.)

- [ ] **Schritt 3: memory-bank via MCP prüfen**

  Tool: `memory_bank_read` mit `file_name: "techContext.md"` — Inhalt sollte Stack-Info enthalten.
  Tool: `memory_bank_read` mit `file_name: "systemPatterns.md"` — Inhalt sollte Struktur enthalten.

- [ ] **Schritt 4: PR erstellen**

  ```bash
  git log main..HEAD --oneline
  ```
  Commits dieser Branch anzeigen, dann PR erstellen.
