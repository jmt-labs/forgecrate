# Active Context

## Aktueller Fokus

Workflow-Overhaul geplant (Spec: `docs/superpowers/specs/2026-06-02-workflow-overhaul-design.md`):
- Hooks warnen nur noch, blockieren nicht mehr
- Neuer `forgecrate-research` Skill (base layer)
- PR-Pflicht-Sequenz: doc-sync → handoff → db-migration → accessibility-audit → ui-ux-audit → pr-checklist
- batch-issues in issue-resolver integriert, getbetter schreibt in memory MCP
- roadmap-triage wird Pflicht nach brainstorming (WSJF + K.O.-Kriterien)

Abgeschlossen:
- PR #114 gemergt: Claude Plugins-Abschnitt in base/CLAUDE.md dokumentiert
- Validierung von Profil-/Flavor-Namen (internal/deploy/validate.go)
- codegraph-Flavor implementiert (Issue #87, #88)

## Offene Fragen

- Sollen andere Repos mit `forgecrate update` die neuen Flavor-Mechanismen automatisch erhalten?

## Bekannte Blocker

Keine. (e2e-Plugin-Install-Fehler bei Frontend-Plugins sind netz-/sandboxbedingt,
unabhängig von der Validierung; `make test-e2e` mit Fake-CLAUDE_BIN ist grün.)
