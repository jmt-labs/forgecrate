# Workflow Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Skills und Hooks in `base/`, `flavors/`, `profiles/` anpassen, damit Ziel-Repos via `forgecrate deploy` den neuen Workflow erhalten: Hooks warnen statt zu blockieren, neuer `forgecrate-research`-Skill, erweiterte PR-Pflicht-Sequenz, Skill-Konsolidierungen.

**Architecture:** Alle Änderungen in den Source-Dateien (`base/`, `flavors/`); diese landen via `forgecrate deploy` in Ziel-Repos. Go-Logik in `cmd/forgecrate/hook.go`: `preToolOutput` und `requireResearchOutput` geben nur noch `additionalContext`-Warnungen aus statt `"continue":false` / `"permissionDecision":"deny"`. SKILL.md-Dateien sind reine Markdown-Texte — kein Build-Schritt nötig.

**Tech Stack:** Go 1.22, Cobra CLI, Shell (bash), Markdown

---

## Dateikarte

| Datei | Änderung |
|---|---|
| `cmd/forgecrate/hook.go` | `preToolOutput` + `requireResearchOutput` auf warn-only umstellen |
| `cmd/forgecrate/hook_test.go` | Tests auf warn-only aktualisieren |
| `base/hooks/pre-tool.sh` | `"continue":false`-Check entfernen |
| `.claude/hooks/pre-tool.sh` | Identisch zu `base/hooks/pre-tool.sh` |
| `base/CLAUDE.md` | Pflicht-Tabelle, PR-Sequenz, Hook-Schutz-Abschnitt aktualisieren |
| `base/skills/forgecrate-research/SKILL.md` | Neu anlegen |
| `flavors/github/skills/forgecrate-issue-resolver/SKILL.md` | Research-Schritt + Batch-Modus |
| `flavors/github/skills/forgecrate-batch-issues/SKILL.md` | Löschen |
| `flavors/getbetter/skills/getbetter/SKILL.md` | Speicherziel → memory MCP |
| `flavors/github/skills/forgecrate-roadmap-triage/SKILL.md` | WSJF + K.O.-Kriterien, Pflicht nach brainstorming |

---

## Task 1: Hook warn-only — Tests zuerst (TDD)

**Files:**
- Modify: `cmd/forgecrate/hook_test.go`

- [ ] **Schritt 1: Tests auf warn-only ändern**

In `hook_test.go` folgende Tests anpassen:

```go
func TestPreToolOutput_EditOnMain_Warned(t *testing.T) {
    out := preToolOutput("main", "Edit", "")
    if strings.Contains(out, `"continue":false`) {
        t.Errorf("must not block on main, got: %s", out)
    }
    if !strings.Contains(out, "Warnung") {
        t.Errorf("expected warning in output, got: %s", out)
    }
}

func TestPreToolOutput_EditOnMain_NotBlocked(t *testing.T) {
    out := preToolOutput("main", "Edit", "")
    if strings.Contains(out, `"continue":false`) {
        t.Errorf("edit on main must produce warning, not block, got: %s", out)
    }
}
```

Den bestehenden `TestPreToolOutput_EditOnMain_Blocked` umbenennen zu `TestPreToolOutput_EditOnMain_Warned` und die Assertion von `contains "continue":false` zu `contains "Warnung"` ändern.

Den bestehenden `TestPreToolOutput_DestructiveBashOnMain_Blocked` zu `TestPreToolOutput_DestructiveBashOnMain_Warned` umbenennen:

```go
func TestPreToolOutput_DestructiveBashOnMain_Warned(t *testing.T) {
    for _, cmd := range []string{"git reset --hard", "git commit -m msg", "git push origin main", "git push --force", "git clean -f"} {
        out := preToolOutput("main", "Bash", cmd)
        if strings.Contains(out, `"continue":false`) {
            t.Errorf("cmd %q on main: must not block (warn only), got: %s", cmd, out)
        }
        if !strings.Contains(out, "Warnung") {
            t.Errorf("cmd %q on main: expected warning, got: %s", cmd, out)
        }
    }
}
```

`TestRequireResearchOutput` — Subtest `"Edit without research blocks"` auf warn-only ändern:

```go
t.Run("Edit without research warns", func(t *testing.T) {
    in := `{"tool_name":"Edit","transcript_path":"` + noResearch + `"}`
    out := requireResearchOutput(strings.NewReader(in), dir)
    if strings.Contains(out, `"permissionDecision":"deny"`) {
        t.Errorf("must not deny, got: %q", out)
    }
    if !strings.Contains(out, "Recherche") {
        t.Errorf("expected research warning, got: %q", out)
    }
})
```

`TestResearchDecision` — alle `wantBlock: true`-Cases auf `wantWarn: true` und Rückgabewert `(bool, string)` anpassen (Name ändert sich später mit Implementierung).

- [ ] **Schritt 2: Tests ausführen — müssen fehlschlagen**

```bash
go test ./cmd/forgecrate/... -run "TestPreToolOutput_EditOnMain|TestPreToolOutput_Destructive|TestRequireResearchOutput" -v
```

Erwartung: FAIL (Assertions noch auf altes Verhalten ausgerichtet).

- [ ] **Schritt 3: Commit**

```bash
git add cmd/forgecrate/hook_test.go
git commit -m "test(hook): warn-only Tests — Block-Assertions auf Warn-Assertions umstellen"
```

---

## Task 2: Hook warn-only — Go-Implementierung

**Files:**
- Modify: `cmd/forgecrate/hook.go`

- [ ] **Schritt 1: `preToolOutput` auf warn-only umstellen**

In `cmd/forgecrate/hook.go` die Funktion `preToolOutput` komplett ersetzen:

```go
func preToolOutput(branch, toolName, toolInput string) string {
    onMain := isMainBranch(branch)

    switch toolName {
    case "Edit", "Write", "MultiEdit":
        if onMain {
            out, _ := json.Marshal(map[string]any{
                "hookSpecificOutput": map[string]string{
                    "hookEventName":     "PreToolUse",
                    "additionalContext": "Warnung: Direkte Änderungen auf main. Branch anlegen empfohlen: git checkout -b feat/<thema>",
                },
            })
            return string(out)
        }
    case "Bash":
        destructive := isDestructiveBash(toolInput)
        if destructive == "" {
            return ""
        }
        msg := "Warnung: destruktiver Befehl erkannt (" + destructive + ")."
        if onMain {
            msg += " Direkt auf main — Branch anlegen empfohlen."
        }
        out, _ := json.Marshal(map[string]any{
            "hookSpecificOutput": map[string]string{
                "hookEventName":     "PreToolUse",
                "additionalContext": msg,
            },
        })
        return string(out)
    }
    return ""
}
```

- [ ] **Schritt 2: `requireResearchOutput` auf warn-only umstellen**

Die Funktion `requireResearchOutput` in `hook.go` ersetzen:

```go
func requireResearchOutput(r io.Reader, dir string) string {
    data, err := io.ReadAll(r)
    if err != nil {
        return ""
    }
    var in preToolInput
    if err := json.Unmarshal(data, &in); err != nil {
        return ""
    }

    cfg, _ := readForgecrateConfig(dir)

    if in.TranscriptPath == "" {
        return ""
    }
    transcript, err := os.ReadFile(in.TranscriptPath)
    if err != nil {
        return ""
    }

    warn, reason := researchDecision(cfg, transcript, in.ToolName, in.ToolInput.Command)
    if !warn {
        return ""
    }
    out, err := json.Marshal(map[string]any{
        "hookSpecificOutput": map[string]string{
            "hookEventName":     "PreToolUse",
            "additionalContext": reason,
        },
    })
    if err != nil {
        return ""
    }
    return string(out)
}
```

- [ ] **Schritt 3: `researchBlockMessage` anpassen**

```go
const researchBlockMessage = "Recherche-Empfehlung: Vor Edit/Write/MultiEdit mindestens ein Recherche-Tool (WebSearch, WebFetch, mcp__fetch__*, mcp__context7__*) nutzen — nicht raten. Danach sind weitere Edits der Session frei. Verzicht via Flavor no-research."
```

- [ ] **Schritt 4: Tests ausführen — müssen bestehen**

```bash
go test ./cmd/forgecrate/... -v
```

Erwartung: alle Tests PASS.

- [ ] **Schritt 5: Alle Tests**

```bash
go test ./...
```

Erwartung: alle Pakete `ok`.

- [ ] **Schritt 6: Commit**

```bash
git add cmd/forgecrate/hook.go
git commit -m "feat(hook): warn-only — preToolOutput und requireResearchOutput blockieren nicht mehr"
```

---

## Task 3: Hook-Shell-Skripte aktualisieren

**Files:**
- Modify: `base/hooks/pre-tool.sh`
- Modify: `.claude/hooks/pre-tool.sh`

- [ ] **Schritt 1: `base/hooks/pre-tool.sh` anpassen**

Den `"continue":false`-Check entfernen. Das Skript braucht keinen frühen Exit mehr:

```bash
#!/usr/bin/env bash
# PreToolUse-Hook. Warnt bei destruktiven Befehlen und fehlender Recherche.
# Blockiert nie — Entscheidung liegt beim Agenten.

STDIN_JSON=""
if [ ! -t 0 ]; then
  STDIN_JSON=$(cat)
fi

if command -v forgecrate >/dev/null 2>&1; then
  # Destruktive-Befehl-Warnung (alle Branches)
  OUT=$(printf '%s' "$STDIN_JSON" | forgecrate hook pre-tool)
  if [ -n "$OUT" ]; then
    printf '%s' "$OUT"
  fi

  # Recherche-Empfehlung: warnt bei Edit/Write/MultiEdit ohne vorherige Recherche
  DECISION=$(printf '%s' "$STDIN_JSON" | forgecrate hook require-research)
  if [ -n "$DECISION" ]; then
    printf '%s' "$DECISION"
  fi
fi
```

- [ ] **Schritt 2: `.claude/hooks/pre-tool.sh` identisch setzen**

Gleicher Inhalt wie `base/hooks/pre-tool.sh`.

- [ ] **Schritt 3: Commit**

```bash
git add base/hooks/pre-tool.sh .claude/hooks/pre-tool.sh
git commit -m "feat(hooks): pre-tool.sh blockiert nicht mehr, nur noch Warnungen"
```

---

## Task 4: `base/CLAUDE.md` Template aktualisieren

**Files:**
- Modify: `base/CLAUDE.md`

- [ ] **Schritt 1: Pflicht-Skills-Tabelle erweitern**

Den Abschnitt `## Pflicht-Skills` in `base/CLAUDE.md` ersetzen:

```markdown
## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Nach Brainstorming | `forgecrate-roadmap-triage` | MUSS aufgerufen werden — entscheidet ob jetzt oder Future Feature |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor jeder nicht-trivialen Änderung | `forgecrate-research` | MUSS aufgerufen werden |
| Vor jeder nicht-trivialen Änderung | Codegraph (`codegraph_node` + `codegraph_callers`) | MUSS für betroffene Symbole ausgeführt werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |
| Session-Start | `mcp__memory__read_graph` | Projektübergreifendes Wissen laden |
| Architekturentscheidung / Debugging-Ergebnis | `mcp__memory__*` | In memory MCP schreiben |
```

- [ ] **Schritt 2: Recherche-Pflicht-Abschnitt aktualisieren**

Den Abschnitt `## Recherche-Pflicht (erzwungen)` ersetzen:

```markdown
## Recherche-Pflicht

**Alle** Rollen MÜSSEN vor jeder nicht-trivialen Code-Änderung mindestens ein
Recherche-Tool nutzen — statt aus gelerntem Wissen zu arbeiten. Raten ist verboten;
Quellen werden referenziert. Der `pre-tool.sh`-Hook **warnt** bei fehlender Recherche,
blockiert aber nicht.

| Frage-Typ | Tool | Beispiele |
|---|---|---|
| Library-/Framework-Doku | `context7` | API-Syntax, Migrationen, Versions-Updates |
| Spezifische URL aus Issue/Ticket | `fetch` MCP | RFCs, MDN, Changelogs |
| Allgemeine Web-Recherche | `WebSearch` | Best Practices, Vergleiche, aktuelle Probleme |

**Regeln:**

- Mindestens eine Quelle pro nicht-trivialer Entscheidung
- Quellen im Plan-Dokument (`docs/superpowers/plans/*.md`) referenzieren
- Deaktivierbar via Flavor `no-research`
```

- [ ] **Schritt 3: Entwicklungs-Workflow — PR-Pflicht-Sequenz ergänzen**

Im Abschnitt `## Entwicklungs-Workflow`, Punkt 5 ersetzen:

```markdown
5. **PR & Abschluss** — Vor `gh pr create` diese Sequenz vollständig ausführen:
   1. `forgecrate-doc-sync` — Doku mit Code abgleichen
   2. `forgecrate-handoff` — memory-bank aktualisieren
   3. `forgecrate-db-migration` — Migrations-Review
   4. `accessibility-audit` — A11y-Prüfung
   5. `ui-ux-audit` — UX-Review
   6. `forgecrate-pr-checklist` — Abschluss-Checkliste

   Dann PR erstellen, Issue im PR-Body verlinken ("Closes #N").
   Issue wird nach Merge automatisch geschlossen.
```

- [ ] **Schritt 4: Hook-Schutz-Abschnitt aktualisieren**

Den Abschnitt `## Hook-Schutz: Hinweis` ersetzen:

```markdown
## Hook-Schutz: Hinweis

Der `pre-tool.sh`-Hook **warnt** bei destruktiven Bash-Befehlen und fehlender Recherche —
er blockiert nie. Die Verantwortung liegt beim Agenten: Warnungen bewusst wahrnehmen,
einschätzen und eine informierte Entscheidung treffen.

Für serverseitigen Schutz auf `main`: GitHub Branch Protection Rules konfigurieren.
```

- [ ] **Schritt 5: Commit**

```bash
git add base/CLAUDE.md
git commit -m "feat(base): CLAUDE.md-Template — warn-only, PR-Pflicht-Sequenz, Research-Skill-Pflicht"
```

---

## Task 5: Neuer Skill `forgecrate-research`

**Files:**
- Create: `base/skills/forgecrate-research/SKILL.md`

- [ ] **Schritt 1: Skill anlegen**

```bash
mkdir -p base/skills/forgecrate-research
```

Inhalt `base/skills/forgecrate-research/SKILL.md`:

```markdown
---
name: forgecrate-research
description: >-
  Pflicht vor jeder nicht-trivialen Änderung. Recherchiert aktuelle Doku,
  Best Practices oder spezifische Quellen und gibt einen strukturierten
  Research-Block aus, den folgende Skills referenzieren können.
  UNBEDINGT aufrufen bevor Code geschrieben oder geändert wird.
---

# Research

Recherchiert das Thema und liefert einen strukturierten Block mit Quellen
und Erkenntnissen — als Grundlage für nachfolgende Implementierung.

## Ablauf

**Schritt 1 — Thema bestimmen:**
Formuliere eine präzise Recherche-Frage. Was genau muss geklärt werden?
Welche Library, API, Pattern oder Entscheidung steht an?

**Schritt 2 — Tool wählen:**

| Frage-Typ | Tool |
|---|---|
| Library-/Framework-Doku (API-Syntax, Migration, Version) | `mcp__context7__*` |
| Spezifische URL (RFC, MDN, Changelog, Issue-Link) | `mcp__fetch__fetch` |
| Allgemeine Fragen (Best Practices, Vergleiche, Alternativen) | `WebSearch` |

Mindestens ein Tool pro Session vor dem ersten Edit/Write nutzen.

**Schritt 3 — Recherche durchführen:**
Tool aufrufen. Bei unklaren Ergebnissen zweites Tool oder verfeinerte Query nutzen.

**Schritt 4 — Research-Block ausgeben:**

```
## Research-Ergebnis

**Frage:** <präzise Frage>
**Quelle(n):** <URL oder Tool + Query>
**Erkenntnisse:**
- <Kernaussage 1>
- <Kernaussage 2>
**Implikation für diese Änderung:** <ein Satz>
```

Diesen Block im Plan-Dokument referenzieren (`docs/superpowers/plans/*.md`).

## Eingebettet in andere Skills

Dieser Skill wird als erster Schritt in folgenden Skills aufgerufen:
- `superpowers:brainstorming` — vor dem Design
- `superpowers:test-driven-development` — vor der Implementierung
- `forgecrate-issue-resolver` — vor der Analyse
```

- [ ] **Schritt 2: Commit**

```bash
git add base/skills/forgecrate-research/
git commit -m "feat(skill): forgecrate-research — neuer Pflicht-Skill vor nicht-trivialen Änderungen"
```

---

## Task 6: `forgecrate-issue-resolver` — Research-Schritt + Batch-Modus

**Files:**
- Modify: `flavors/github/skills/forgecrate-issue-resolver/SKILL.md`
- Delete: `flavors/github/skills/forgecrate-batch-issues/SKILL.md`

- [ ] **Schritt 1: Research-Schritt als ersten Block einbauen**

In `flavors/github/skills/forgecrate-issue-resolver/SKILL.md` nach der Einleitung und vor `### 1. Verstehen` einen neuen Block einfügen:

```markdown
### 0. Research (`forgecrate-research`)
Bevor das Issue analysiert wird: `forgecrate-research` aufrufen.
Thema: betroffene Technologie, Muster oder API aus dem Issue-Kontext.
Research-Block im Issue-Kommentar `🔍 Research` dokumentieren.
```

- [ ] **Schritt 2: Batch-Modus-Abschnitt ergänzen**

Am Anfang von `forgecrate-issue-resolver/SKILL.md` (nach dem YAML-Header) den Abschnitt `## Argumente` hinzufügen:

```markdown
## Argumente (`$ARGUMENTS`)

Alle optional, frei kombinierbar:

- Einzelne Issue-Nummer (z. B. `42`) — Standard-Modus, ein Issue
- `count:N` — N Issues parallel bearbeiten (Default 1, Maximum 5). Auto-Auswahl der nächsten offenen Issues nach Priorität.
- `label:<name>` — nur Issues mit diesem Label berücksichtigen
- Mehrere Issue-Nummern (z. B. `12 15 18`) — diese Issues parallel bearbeiten

Bei `count > 1` oder mehreren Issue-Nummern:
1. Issues auswählen/bestätigen
2. Pro Issue einen isolierten Subagenten via `isolation: "worktree"` + `run_in_background: true` dispatchen
3. Jeder Subagent führt den vollständigen Issue-Resolver-Workflow aus
4. Koordination: Fortschritt wird je Issue als Issue-Kommentar dokumentiert
```

- [ ] **Schritt 3: `forgecrate-batch-issues` löschen**

```bash
rm -rf flavors/github/skills/forgecrate-batch-issues
```

- [ ] **Schritt 4: Commit**

```bash
git add flavors/github/skills/forgecrate-issue-resolver/SKILL.md
git rm -r flavors/github/skills/forgecrate-batch-issues/
git commit -m "feat(skill): issue-resolver — Research-Schritt + Batch-Modus; batch-issues Skill entfernt"
```

---

## Task 7: `getbetter` → memory MCP

**Files:**
- Modify: `flavors/getbetter/skills/getbetter/SKILL.md`

- [ ] **Schritt 1: SKILL.md lesen**

```bash
cat flavors/getbetter/skills/getbetter/SKILL.md
```

- [ ] **Schritt 2: Speicherziel auf memory MCP umstellen**

Im Ablauf-Abschnitt des Skills den letzten Schritt ersetzen. Statt in `.claude/GETBETTER.md` zu schreiben:

```markdown
**Letzter Schritt — In memory MCP speichern:**

`mcp__memory__add_observations` aufrufen mit:
- `entityName`: `session-reflection`
- `contents`: Liste der synthetisierten Erkenntnisse (je Erkenntnis ein String)

Format pro Erkenntnis:
```
[YYYY-MM-DD] <Kategorie>: <Erkenntnis in einem Satz>
```

Kategorien: `workflow`, `tooling`, `pattern`, `mistake`, `decision`.

Kein GETBETTER.md schreiben — memory MCP ist das einzige Speicherziel.
```

- [ ] **Schritt 3: Commit**

```bash
git add flavors/getbetter/skills/getbetter/SKILL.md
git commit -m "feat(skill): getbetter — Speicherziel von GETBETTER.md auf memory MCP umgestellt"
```

---

## Task 8: `forgecrate-roadmap-triage` — Pflicht nach brainstorming + WSJF/K.O.

**Files:**
- Modify: `flavors/github/skills/forgecrate-roadmap-triage/SKILL.md`

- [ ] **Schritt 1: Description und Trigger aktualisieren**

YAML-Header anpassen:

```yaml
---
name: forgecrate-roadmap-triage
description: >-
  PFLICHT nach jedem abgeschlossenen Brainstorming. Bewertet das Ergebnis
  mit WSJF-Score und K.O.-Kriterien. Entscheidet: "Umsetzbar jetzt" (→ writing-plans)
  oder "Future Feature" (→ GitHub Issue anlegen). Auch aufrufbar wenn neue Ideen,
  Bugs oder Feature-Wünsche erwähnt werden.
---
```

- [ ] **Schritt 2: Entscheidungslogik einbauen**

Nach dem bestehenden WSJF-Scoring-Abschnitt einen neuen Abschnitt `## Entscheidung` einfügen:

```markdown
## Entscheidung

### K.O.-Kriterien (überwiegen WSJF-Score)

Falls eines zutrifft → sofort **Future Feature**, unabhängig vom Score:

1. **Fehlende Abhängigkeiten** — externes Service, anderes Feature oder Infrastruktur noch nicht fertig
2. **Scope zu groß** — nicht in einem PR + Review-Zyklus abschließbar (>3 Tage Arbeit)
3. **Kein definierbarer Nutzer-Impact** — Akzeptanzkriterien nicht formulierbar

### WSJF-Schwelle

- Score ≥ 2.0 → **Umsetzbar jetzt**
- Score < 2.0 → **Future Feature**

### Ausgabe

**Umsetzbar jetzt:**
```
✅ Umsetzbar jetzt (WSJF: X.X)
→ Weiter mit writing-plans
```

**Future Feature:**
```
📋 Future Feature (WSJF: X.X | K.O.: <Grund>)
→ GitHub Issue anlegen mit Label "future-feature"
→ Kein Plan, keine Implementierung jetzt
```

## Trigger

Wird aufgerufen:
- **Pflicht** nach jedem abgeschlossenen `superpowers:brainstorming`
- Wenn der Nutzer eine neue Idee, einen Bug, eine Erweiterung oder Feature-Wunsch erwähnt
- Wenn gefragt wird "ist das wichtig?" / "gehört das auf die Roadmap?" / "machen wir das?"
```

- [ ] **Schritt 3: Commit**

```bash
git add flavors/github/skills/forgecrate-roadmap-triage/SKILL.md
git commit -m "feat(skill): roadmap-triage — Pflicht nach brainstorming, WSJF + K.O.-Kriterien"
```

---

## Task 9: Verifikation

- [ ] **Schritt 1: Alle Tests**

```bash
go test ./...
```

Erwartung: alle Pakete `ok`.

- [ ] **Schritt 2: Quality**

```bash
make quality
```

Erwartung: keine Fehler.

- [ ] **Schritt 3: Geänderte Skill-Dateien prüfen**

```bash
ls base/skills/forgecrate-research/
ls -la flavors/github/skills/ | grep batch   # muss leer sein
head -5 flavors/getbetter/skills/getbetter/SKILL.md
head -20 flavors/github/skills/forgecrate-roadmap-triage/SKILL.md
```

- [ ] **Schritt 4: Hook-Verhalten manuell testen**

```bash
bash .claude/hooks/pre-tool-test.sh
```

Erwartung: Script läuft durch, keine `"continue":false`-Outputs.

- [ ] **Schritt 5: Abschluss-Commit (falls nötig)**

```bash
git status
# Nur committen wenn noch unstaged changes
```
