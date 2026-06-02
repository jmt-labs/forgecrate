# Workflow Overhaul Design

**Datum:** 2026-06-02
**Scope:** Hook-Verhalten, Pflicht-Skills, Skill-Konsolidierungen, Research-Skill

---

## Kontext

Der bisherige Workflow hatte zwei Schwachstellen:
1. Hooks blockierten den Arbeitsablauf hart — führte zu Stillstand statt bewusster Entscheidung
2. Pflichten waren über Hook-Logik und CLAUDE.md verteilt, nicht konsistent in Skills verankert

Dieser Overhaul verschiebt die Verantwortung vollständig in Skills und CLAUDE.md (Ansatz A: Flat list), Hooks warnen nur noch.

---

## 1. Hook-Verhalten

**Änderung:** Beide Checks in `pre-tool.sh` (`pre-tool` + `require-research`) geben nicht mehr `"continue": false` zurück. Sie liefern nur noch eine Warnung als Nachricht und lassen das Tool passieren.

**Begründung:** Bewusstsein statt Zwang. Der Agent soll Probleme wahrnehmen, einschätzen und eine bewusste Entscheidung treffen — nicht automatisch gestoppt werden.

**Betroffene Dateien:**
- `.claude/hooks/pre-tool.sh` — Ausgabe-Format ändern: kein `"continue": false`, nur Warn-Text
- `base/hooks/pre-tool.sh` — identisch
- `internal/hook/` — `pre-tool` und `require-research` Subkommandos: Exit-Code und JSON-Output anpassen

`pre-tool-test.sh` bleibt unverändert.

---

## 2. Neuer Skill: `forgecrate-research`

**Ort:** `base/skills/forgecrate-research/SKILL.md`

**Zwei Nutzungskontexte:**

### Eigenständig
Aufrufbar als `/forgecrate-research`. Ablauf:
1. Frage/Thema formulieren
2. Tool wählen: context7 (Library-Doku) / WebSearch (Best Practices, Vergleiche) / fetch (spezifische URL)
3. Quellen zusammenfassen
4. Strukturierten Research-Block ausgeben, den folgende Skills referenzieren

### Eingebettet
Die Skills `brainstorming`, `superpowers:test-driven-development` und `forgecrate-issue-resolver` bekommen einen verpflichtenden ersten Block „## Research" der `forgecrate-research` aufruft, bevor Code oder Plan entsteht.

**CLAUDE.md-Eintrag:** `forgecrate-research` als Pflicht-Skill vor jeder nicht-trivialen Änderung — in der Pflicht-Tabelle analog zu `brainstorming`.

---

## 3. Pflicht-Skills vor PR

In CLAUDE.md wird der PR-Workflow-Abschnitt um eine feste Sequenz erweitert. Alle Skills laufen bedingungslos vor jedem `gh pr create`:

| Reihenfolge | Skill | Zweck |
|---|---|---|
| 1 | `forgecrate-doc-sync` | Doku mit aktuellem Code abgleichen |
| 2 | `forgecrate-handoff` | memory-bank mit Session-Kontext aktualisieren |
| 3 | `forgecrate-db-migration` | Migrations-Review |
| 4 | `accessibility-audit` | A11y-Prüfung geänderter UI-Komponenten |
| 5 | `ui-ux-audit` | Ganzheitlicher UX-Review |
| 6 | `forgecrate-pr-checklist` | Abschluss-Checkliste (bereits Pflicht) |

### Codegraph-Pflicht
`codegraph_node` + `codegraph_callers` für betroffene Symbole — Pflicht vor jeder nicht-trivialen Änderung. Als eigene Zeile in die CLAUDE.md-Pflicht-Tabelle.

### Memory MCP-Pflicht
- Lesen: am Session-Start via `mcp__memory__*`
- Schreiben: bei Architekturentscheidungen, Debugging-Ergebnissen, Brainstorming-Ergebnissen

Als eigene Zeile in die CLAUDE.md-Pflicht-Tabelle.

---

## 4. Skill-Konsolidierungen

### `forgecrate-batch-issues` → in `forgecrate-issue-resolver` integriert

`forgecrate-issue-resolver` bekommt optionale Argumente:
- `count:N` — N Issues parallel (max 5, default 1)
- `label:<name>` — Filterung nach Label
- explizite Issue-Nummern — haben Vorrang vor Auto-Auswahl

Bei `count > 1` oder mehreren Issue-Nummern orchestriert der issue-resolver selbst parallele Subagenten — einen pro Issue. `forgecrate-batch-issues` als eigenständiger Skill wird gelöscht.

### `getbetter` — Speicherziel wechselt zu memory MCP

Skill-Logik und Trigger bleiben unverändert. Statt in `.claude/GETBETTER.md` zu schreiben, ruft der Skill am Ende `mcp__memory__*` auf und speichert Session-Erkenntnisse als strukturierten Eintrag. `.claude/GETBETTER.md` entfällt.

### `forgecrate-roadmap-triage` — Pflicht nach Brainstorming, neue Entscheidungslogik

**Trigger:** Pflicht nach jedem abgeschlossenen Brainstorming (in CLAUDE.md und im brainstorming-Skill als nächster Schritt vor writing-plans).

**Entscheidungslogik:**
1. WSJF-Score berechnen (Cost of Delay ÷ Job Size, bestehende Skala)
2. K.O.-Kriterien prüfen — führen unabhängig vom Score direkt zu „Future Feature":
   - Fehlende Abhängigkeiten (externe Services, andere Features nicht fertig)
   - Scope zu groß für einen PR (nicht in einem Review-Zyklus abschließbar)
   - Kein klar definierbarer Nutzer-Impact

**Ausgabe:**
- `Umsetzbar jetzt` → weiter zu `writing-plans`
- `Future Feature` → GitHub Issue anlegen mit Triage-Begründung, kein Plan

---

## Akzeptanzkriterien

- [ ] `pre-tool.sh` blockiert nie mehr, gibt nur noch Warn-Text aus
- [ ] `forgecrate-research` Skill existiert in `base/skills/`
- [ ] brainstorming, tdd, issue-resolver rufen research als ersten Schritt auf
- [ ] CLAUDE.md enthält vollständige PR-Pflicht-Sequenz (6 Skills)
- [ ] CLAUDE.md enthält Codegraph- und memory MCP-Pflicht
- [ ] `forgecrate-batch-issues` ist gelöscht, issue-resolver unterstützt `count:N`
- [ ] `getbetter` schreibt in memory MCP statt GETBETTER.md
- [ ] `forgecrate-roadmap-triage` hat WSJF + K.O.-Logik und wird nach brainstorming aufgerufen
- [ ] `go test ./...` bleibt grün
