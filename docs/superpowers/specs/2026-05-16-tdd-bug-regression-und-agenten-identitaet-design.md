# Design: TDD Bug-Regression & Agenten-Identität

**Datum:** 2026-05-16  
**Status:** Draft

## Überblick

Zwei unabhängige Ergänzungen zur bestehenden forgecrate-Konfiguration:

1. **TDD Bug-Regression** — Jeder gefundene Bug muss durch einen Regressionstest abgedeckt werden, bevor der Fix committed wird.
2. **Agenten-Identität** — Subagenten erhalten eindeutige Namen und Farben.

---

## Feature 1: TDD Bug-Regression

### Motivation

Das TDD-Flavor erzwingt bisher Tests für neue Features. Bugfixes fallen durch dieses Raster — ein Bug kann gefixt und committed werden, ohne dass ein Test verhindert, dass er wieder auftaucht.

### Änderungen

**`flavors/tdd/CLAUDE.md`** — neue Regel im TDD-Abschnitt:

```
- Jeder gefundene Bug erhält vor dem Fix einen Regressionstest
```

**`base/CLAUDE.md`** — neue Zeile in der Pflicht-Skills-Tabelle:

| Situation | Skill | Verhalten |
|---|---|---|
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |

### Verhalten

1. `superpowers:systematic-debugging` identifiziert den Bug.
2. Danach: `superpowers:test-driven-development` aufrufen — Regressionstest schreiben, der den Bug reproduziert (muss fehlschlagen).
3. Fix implementieren — Test muss bestehen.
4. Erst dann committen.

---

## Feature 2: Agenten-Identität

### Motivation

Beim parallelen Einsatz mehrerer Subagenten ist es schwer nachzuvollziehen, welcher Agent welche Ausgabe produziert. Eindeutige Namen und Farben schaffen Orientierung im FleetView.

### Änderungen

**`base/CLAUDE.md`** — neue Untersektion "Agenten-Identität" unter "Parallelisierung & Isolation":

```markdown
### Agenten-Identität

Jeder Subagent bekommt:
- **Eindeutigen Namen** — via `description`-Parameter im Agent-Tool-Aufruf (3–5 Wörter, Rolle + Aufgabe)
- **Eindeutige Farbe** — dynamisch durch FleetView zugewiesen; keine zwei gleichzeitig laufenden Agenten teilen eine Farbe
```

### Verhalten

- `description` beschreibt Rolle und Aufgabe knapp (z.B. "Analyst: Auth-Flow", "Reviewer: Deploy-Code").
- Farben werden nicht manuell gesetzt — FleetView weist sie dynamisch zu.
- Bei mehreren gleichzeitigen Agenten: vor dem Dispatch sicherstellen, dass jeder `description` gesetzt hat.

---

## Nicht im Scope

- Keine Änderung an Skills selbst (nur CLAUDE.md-Regeln).
- Keine programmatische Farb-Zuweisung — FleetView übernimmt das.
- Kein neuer Skill für Regressionstests — die bestehende TDD-Skill-Verknüpfung reicht.
