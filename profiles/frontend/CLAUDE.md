## Frontend-Profil

- Komponenten: klein, fokussiert, eine Verantwortlichkeit
- State: lokal wenn möglich, global nur wenn nötig
- Kein CSS-in-JS ohne explizite Anforderung
- Barrierefreiheit: semantisches HTML, ARIA-Attribute wo nötig
- Tests: Behavior-Tests (was der Nutzer sieht), keine Implementierungsdetails

## UI-Reviews

- **`accessibility-audit`** — schnelle statische A11y-Checks pro geänderter Datei (alt, label, aria-*). Eignet sich für Pre-Commit / PR-Reviews.
- **`ui-ux-audit`** — tiefgehender Audit der gesamten UI, gruppiert nach Bereichen, mit Severity-Bewertung und automatischer Erstellung kleinteiliger GitHub-Issues. Für Major-Releases oder größere UI-Refactorings.

## Playwright MCP

Browser-Automatisierung direkt aus Claude heraus. Automatisch konfiguriert via `profiles/frontend/extensions.yaml`.

**Verwende es für:** UI-Tests, Screenshots, Formular-Interaktionen, visuelle Regressionstests, Debugging von Rendering-Problemen.

**Verwende es NICHT für:** API-Tests ohne UI-Beteiligung (→ direkte HTTP-Calls), GitHub-Operationen (→ github MCP).

## Design-Plugins

Fünf spezialisierte Plugins für UI/UX-Arbeit — optimal in diesen Situationen:

| Plugin | Optimal wenn… |
|---|---|
| `ui-ux-pro-max-skill` | Neue Komponente/Seite designen — generiert automatisch Design-System (Farben, Typografie, Spacing) passend zum Produkt; unterstützt React, Next.js, Vue, Tailwind, Flutter u.v.m. |
| `interface-design` | UI über mehrere Sessions konsistent halten — speichert Design-Entscheidungen (Spacing, Elevation, Farben) in `.interface-design/system.md` und wendet sie session-übergreifend an |
| `refactoring-ui-skill` | Bestehende UI überarbeiten — `/ui-refactor` verbessert Hierarchie, Spacing (8px-Raster), HSL-Farben und Schatten nach Refactoring-UI-Prinzipien |
| `agent-skills` | Vercel-Deployments oder React Composition Patterns — auto-detects 40+ Frameworks, hilft bei Compound Components, State-Lifting und Edge-Funktionen |
| `wondelai-skills` | UX-Strategie und Produktentscheidungen — 25 Skills nach Norman, Cialdini, Ries; deckt UX Design, Conversion-Optimierung und Produktstrategie ab |
