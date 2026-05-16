## Fullstack-Profil

Kombiniert Backend- und Frontend-Anforderungen.

- API-Kontrakte explizit definieren bevor Implementierung auf beiden Seiten
- Shared Types: einmal definieren, in beiden Schichten nutzen
- End-to-End-Tests für kritische User-Flows

## Playwright MCP

Browser-Automatisierung direkt aus Claude heraus. Automatisch konfiguriert via `profiles/fullstack/extensions.yaml`.

**Verwende es für:** E2E-Tests über Frontend und API hinweg, Screenshots, Formular-Interaktionen, visuelle Regressionstests.

**Verwende es NICHT für:** reine API-Tests ohne UI-Beteiligung (→ direkte HTTP-Calls), GitHub-Operationen (→ github MCP).
