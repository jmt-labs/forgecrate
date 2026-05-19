## GitHub-Flavor

- Releases über `gh release create` veröffentlichen (nach `release`-Skill)
- PR-Templates in `.github/pull_request_template.md` pflegen
- CI-Status mit `gh run list` prüfen bevor ein Release getaggt wird

## Multiagent & Subagenten

Bei jeder Aufgabe im GitHub-Kontext proaktiv prüfen:
- Task >1 min oder Ergebnis nicht sofort nötig → `run_in_background: true`
- Feature-Branch, Multi-File-Änderung, langer Plan → `isolation: "worktree"`
- Mehrere unabhängige Tasks gleichzeitig → beides kombinieren

Warten ist kein Default — im Zweifelsfall Background nutzen.
