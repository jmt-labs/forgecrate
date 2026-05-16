## Frontend-Profil

- Komponenten: klein, fokussiert, eine Verantwortlichkeit
- State: lokal wenn möglich, global nur wenn nötig
- Kein CSS-in-JS ohne explizite Anforderung
- Barrierefreiheit: semantisches HTML, ARIA-Attribute wo nötig
- Tests: Behavior-Tests (was der Nutzer sieht), keine Implementierungsdetails

## Playwright MCP

Browser-Automatisierung direkt aus Claude heraus. Automatisch konfiguriert via `profiles/frontend/extensions.yaml`.

**Verwende es für:** UI-Tests, Screenshots, Formular-Interaktionen, visuelle Regressionstests, Debugging von Rendering-Problemen.

**Verwende es NICHT für:** API-Tests ohne UI-Beteiligung (→ direkte HTTP-Calls), GitHub-Operationen (→ github MCP).
