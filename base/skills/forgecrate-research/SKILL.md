---
name: forgecrate-research
description: >-
  PFLICHT vor jeder nicht-trivialen Änderung. Recherchiert aktuelle Doku,
  Best Practices oder spezifische Quellen und gibt einen strukturierten
  Research-Block aus, den folgende Skills referenzieren können.
  UNBEDINGT aufrufen bevor Code geschrieben oder geändert wird.
---

# Research

Recherchiert das Thema und liefert einen strukturierten Block mit Quellen und
Erkenntnissen — als Grundlage für nachfolgende Implementierung.

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
