# Active Context

## Aktueller Fokus

Validierung von Profil-/Flavor-Namen ergänzt (Branch
`claude/project-improvement-planning-JVwWc`):
- Tippfehler wie `--profile backendd` / `--flavors tddd` wurden bisher in
  `compose` still übersprungen (`if err == nil`) → unvollständige Konfig ohne Fehler.
- Neu: `deploy.validateSelection` (internal/deploy/validate.go) prüft gegen den
  tatsächlichen Katalog (profiles/, flavors/) und bricht VOR jedem Schreibvorgang
  mit klarer Meldung + Levenshtein-„meintest du …?"-Vorschlag ab.
- Choke-Point: `deploy.RunWithClaude` (deckt init/update/config zugleich ab).
- Regel: leerer/fehlender Katalog → Prüfung entfällt (für Minimal-Test-Fixtures).

## Offene Fragen

- Sollen andere Repos mit `forgecrate update` die neue Workflow-Regel automatisch erhalten?

## Bekannte Blocker

Keine. (e2e-Plugin-Install-Fehler bei Frontend-Plugins sind netz-/sandboxbedingt,
unabhängig von der Validierung; `make test-e2e` mit Fake-CLAUDE_BIN ist grün.)
