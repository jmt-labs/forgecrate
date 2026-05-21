# Design: memory-bank als einzige Projektkontext-Quelle

**Datum:** 2026-05-21  
**Status:** Approved

## Problem

Projektspezifische Inhalte (Sprache, Struktur, Workflow, Abhängigkeiten) werden aktuell an zwei Orten gepflegt:

1. Im `GENERATED`-Block von `CLAUDE.md` — inline, überschreibbar, nicht für Claude direkt adressierbar
2. In `memory-bank/` — strukturiert, via MCP zugreifbar, für alle Sessions persistent

Das führt zu Doppelpflege und Inkonsistenz. CLAUDE.md sollte ausschließlich verhaltensteuernde Inhalte enthalten (Skills, Workflow, Teams). Projektkontext gehört in die memory-bank.

## Ziel

- `CLAUDE.md` enthält keinen projektspezifischen Inline-Inhalt mehr
- `CLAUDE.md` verweist auf den memory-bank MCP-Server als Quelle für Projektkontext
- `forgecrate-repo-onboarding` schreibt ausschließlich in die memory-bank
- `forgecrate-handoff` schreibt ausschließlich in die memory-bank (kein `HANDOFF.md` mehr)
- Migrationspfad für bestehende Repos ist dokumentiert

## Änderungen

### 1. `base/CLAUDE.md` — GENERATED-Block

**Vorher:** Enthält Abschnitte "Projekt-Übersicht", "Struktur", "Workflow", "Externe Abhängigkeiten" als generierten Inline-Inhalt.

**Nachher:** Diese Abschnitte entfallen. Stattdessen ein einziger statischer Hinweis im GENERATED-Block:

```markdown
## Projektkontext

Nutze den `memory-bank` MCP-Server um den aktuellen Projektkontext zu lesen.
```

Kein Verweis auf Dateinamen — nur auf den MCP-Server.

### 2. `base/skills/forgecrate-repo-onboarding/SKILL.md`

**Vorher:** Schritte 1–4 analysieren das Repo; Schritt 5 erstellt einen CLAUDE.md-Vorschlag; Schritt 6 fragt den Nutzer ob CLAUDE.md ersetzt werden soll; Schritt 7 befüllt memory-bank.

**Nachher:** Schritte 5 und 6 entfallen vollständig. Der Skill analysiert das Repo (Schritte 1–4) und schreibt direkt in die memory-bank via MCP (ehemaliger Schritt 7, jetzt Schritt 5). Keine Rückfrage, kein CLAUDE.md-Vorschlag.

Ausgabe an den Nutzer: Kurze Bestätigung welche memory-bank-Dateien befüllt wurden.

### 3. `base/skills/forgecrate-handoff/SKILL.md`

**Vorher:** Sammelt Repo-Daten (Git, Stack, Struktur, TODOs) und schreibt `HANDOFF.md`.

**Nachher:** Sammelt dieselben Daten, schreibt aber über den memory-bank MCP in:

- `activeContext.md` — aktueller Branch, offene Änderungen, offene Fragen
- `progress.md` — letzte 5–10 Commits als "Recent Activity", TODO/FIXMEs als Known Issues

Kein `HANDOFF.md` wird erzeugt. Ausgabe an den Nutzer: Bestätigung der aktualisierten memory-bank-Dateien.

## Migrationspfad

Bestehende Repos, die bereits Inhalt im GENERATED-Block von `CLAUDE.md` haben:

1. **`forgecrate update` ausführen** — Der GENERATED-Block wird durch den neuen MCP-Verweis ersetzt. Der alte projektspezifische Inhalt ist danach nicht mehr in CLAUDE.md.

2. **memory-bank befüllen** — Zwei Optionen:
   - **Automatisch:** `forgecrate-repo-onboarding` ausführen — analysiert das Repo neu und befüllt memory-bank.
   - **Manuell:** Inhalte aus dem alten GENERATED-Block (vor dem Update gesichert oder aus Git-History) in die entsprechenden memory-bank-Dateien übertragen.

3. **Prüfen:** `memory_bank_read` im nächsten Session-Start bestätigt dass der Kontext verfügbar ist.

## Nicht im Scope

- `revise-claude-md` / `claude-md-management` — externes Plugin, wird nicht geändert
- memory-bank-Dateistruktur — bleibt unverändert
- `forgecrate-advisor` — kein CLAUDE.md-Schreibzugriff, nicht betroffen
