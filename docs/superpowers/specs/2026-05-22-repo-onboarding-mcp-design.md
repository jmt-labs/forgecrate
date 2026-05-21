# Design: forgecrate-repo-onboarding — MCP für memory-bank nutzen

## Kontext

Der Skill `forgecrate-repo-onboarding` schreibt nach der Repo-Analyse Kontext in drei memory-bank-Dateien (`projectbrief.md`, `techContext.md`, `systemPatterns.md`). Bisher wird dafür die Anweisung "Read/Write-Tools direkt" verwendet — was bedeutet, Claude nutzt die nativen Datei-Tools, statt den deklarierten MCP-Server.

## Ziel

Schritt 5 des Skills explizit auf den `memory-bank` MCP-Server umstellen:
- `mcp__memory-bank__memory_bank_read` — vor dem Schreiben lesen, um bestehende Inhalte zu kennen
- `mcp__memory-bank__memory_bank_write` — neue Dateien erstellen oder vollständig ersetzen
- `mcp__memory-bank__memory_bank_update` — bestehende Dateien gezielt ergänzen

## Nicht-Ziele

- Kein Änderung an Schritten 1–4 (Analyse-Phase)
- Kein Änderung an `activeContext.md` / `progress.md` (bleiben leer)
- Keine neue Logik — nur Tool-Wechsel

## Änderung

**Datei:** `base/skills/forgecrate-repo-onboarding/SKILL.md` (Quelle)  
**Datei:** `.claude/skills/forgecrate-repo-onboarding/SKILL.md` (deployed, synchron)

**Alt (Schritt 5, letzter Satz):**
> Schreibe die Dateien mit den Read/Write-Tools direkt.

**Neu (Schritt 5, ersetzte Passage):**
> Lies jede Zieldatei zuerst via `mcp__memory-bank__memory_bank_read`, um bestehende Inhalte zu kennen. Schreibe neue Inhalte via `mcp__memory-bank__memory_bank_write` (vollständiger Ersatz) oder `mcp__memory-bank__memory_bank_update` (gezielte Ergänzung). Verwende keine Read/Write-Datei-Tools für memory-bank-Operationen.
