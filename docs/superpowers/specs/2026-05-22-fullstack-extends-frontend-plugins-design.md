# Design: Fullstack-Vererbung und Frontend-Plugins

**Datum:** 2026-05-22
**Status:** Approved

## Kontext

Zwei unabhängige Verbesserungen am Profil-System:

1. **Fullstack-Profil** soll Backend + Frontend automatisch erben, anstatt deren Inhalte doppelt zu pflegen.
2. **Frontend-Profil** bekommt fünf neue Plugins mit klar definierten Nutzungssituationen.

---

## Teil 1: Fullstack-Vererbung via `extends`

### Problem

Das Fullstack-Profil dupliziert aktuell Backend- und Frontend-Inhalte manuell. Ändert sich eines der Basis-Profile, muss Fullstack manuell nachgezogen werden.

### Lösung: `profile.yaml` mit `extends`-Feld

Neue optionale Datei pro Profil: `profiles/<profile>/profile.yaml`

```yaml
# profiles/fullstack/profile.yaml
extends:
  - backend
  - frontend
```

`compose.go` liest dieses File beim Start und expandiert die Layer-Reihenfolge:

```
base → backend → frontend → fullstack
```

Gilt für alle drei Compose-Pfade:
- `collectMarkdownLayers` (CLAUDE.md, AGENTS.md)
- `composeSkills` (Slash-Commands)
- Extensions-Merge im Deploy-Pfad

### Fullstack-Inhalt nach der Änderung

`profiles/fullstack/CLAUDE.md` behält nur fullstack-spezifische Ergänzungen:
- API-Kontrakte explizit definieren vor Implementierung auf beiden Seiten
- Shared Types: einmal definieren, in beiden Schichten nutzen
- End-to-End-Tests für kritische User-Flows
- Playwright MCP für E2E-Tests

`profiles/fullstack/extensions.yaml` behält nur Playwright MCP — Backend- und Frontend-Extensions werden durch Vererbung eingebunden.

### Neue Funktion in compose.go

```go
type ProfileConfig struct {
    Extends []string `yaml:"extends"`
}

func loadProfileConfig(sourceDir, profile string) ProfileConfig
```

Gibt leere `ProfileConfig` zurück wenn keine `profile.yaml` existiert (backward-compatible).

---

## Teil 2: Fünf neue Frontend-Plugins

### Erweiterung Plugin-Typ

Neues optionales `method`-Feld:

```go
type Plugin struct {
    Name   string `yaml:"name"`
    Source string `yaml:"source"`
    Method string `yaml:"method"` // "marketplace" (default: "") → "install"
}
```

In `install.go`:
- `method: marketplace` → `claude plugin marketplace add <source>`
- alles andere → `claude plugin install --scope project <source>` (bisheriges Verhalten)

### Plugins

| Name | Source | Methode |
|---|---|---|
| `ui-ux-pro-max-skill` | `nextlevelbuilder/ui-ux-pro-max-skill` | marketplace |
| `interface-design` | `Dammyjay93/interface-design` | marketplace |
| `agent-skills` | `vercel-labs/agent-skills` | marketplace |
| `wondelai-skills` | `wondelai/skills` | marketplace |
| `refactoring-ui-skill` | `https://github.com/LovroPodobnik/refactoring-ui-skill` | install |

### Nutzungssituationen (CLAUDE.md)

| Plugin | Optimal wenn… |
|---|---|
| `ui-ux-pro-max-skill` | Neue Komponente/Seite designen — generiert automatisch Design-System (Farben, Typografie, Spacing) passend zum Produkt; unterstützt React, Next.js, Vue, Tailwind, Flutter u.v.m. |
| `interface-design` | UI über mehrere Sessions konsistent halten — speichert Design-Entscheidungen (Spacing, Elevation, Farben) in `.interface-design/system.md` und wendet sie session-übergreifend an |
| `refactoring-ui-skill` | Bestehende UI überarbeiten — `/ui-refactor` verbessert Hierarchie, Spacing (8px-Raster), HSL-Farben und Schatten nach Refactoring-UI-Prinzipien |
| `agent-skills` | Vercel-Deployments oder React Composition Patterns — auto-detects 40+ Frameworks, hilft bei Compound Components, State-Lifting und Edge-Funktionen |
| `wondelai-skills` | UX-Strategie und Produktentscheidungen — 25 Skills nach Norman, Cialdini, Ries; deckt UX Design, Conversion-Optimierung und Produktstrategie ab |

---

## Nicht im Scope

- Kein neues Profil oder Flavor
- Keine Änderung am Base-Layer
- Kein Änderung am `extends`-Verhalten für Flavors (nur Profile)
- Keine automatische Deduplizierung bei Konflikten zwischen geerbten Profilen
