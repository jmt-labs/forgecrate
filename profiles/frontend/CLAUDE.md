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
