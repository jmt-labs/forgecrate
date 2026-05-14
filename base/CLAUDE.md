<!-- GENERATED:BEGIN -->
# Claude-Konfiguration

Dieses Repository verwendet ein reproduzierbares Claude-Setup.
Die generierten Abschnitte dieser Datei werden bei `claude-setup update` überschrieben.
Eigene Anpassungen gehören in den CUSTOM-Abschnitt.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |

## Entwicklungs-Workflow

Für alle Features, Bugfixes und Änderungen:

1. **Brainstorming** — `superpowers:brainstorming` aufrufen, Design abstimmen
2. **Spec** — Branch anlegen (`git checkout -b feat/<thema>`); Spec in `docs/superpowers/specs/YYYY-MM-DD-<thema>-design.md` schreiben und committen; GitHub-Issue anlegen oder verlinken; Branch-Name im Issue vermerken; Kommentar im Issue: "Spec fertig"
3. **Plan** — in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` schreiben und committen; Plan-Pfad im Issue ergänzen; Kommentar: "Plan fertig"
4. **Implementierung** — nach jedem Task kurzer Kommentar im Issue
5. **PR & Abschluss** — PR erstellen, Issue im PR-Body verlinken ("Closes #N"); Issue wird erst nach Merge des PR geschlossen (GitHub macht das automatisch)

Ticket-Kommentare immer kurz (ein Satz): Fortschritt, Pfad, oder Ergebnis.

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer über Branch + PR, nie direkt auf `main`

## Team-Rollen & Subagent-Konfiguration

Der Hauptagent koordiniert als Team-Lead. Subagenten übernehmen Rollen entsprechend ihrer Aufgabe.
Der Hauptagent kann bei Bedarf eigenständig von diesen Empfehlungen abweichen.

| Rolle | Superpowers-Skill | Modell | Effort |
|---|---|---|---|
| Analyst / Product Owner | `superpowers:brainstorming` | `claude-opus-4-7` | high |
| Tech Lead / Architekt | `superpowers:writing-plans` | `claude-opus-4-7` | high |
| Entwickler | `superpowers:test-driven-development` | `claude-sonnet-4-6` | medium |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `claude-haiku-4-5-20251001` | low |
| Reviewer | `superpowers:requesting-code-review` | `claude-sonnet-4-6` | medium |
| QA / Abschluss | `superpowers:verification-before-completion` | `claude-sonnet-4-6` | medium |
| Debugger | `superpowers:systematic-debugging` | `claude-sonnet-4-6` | medium |
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->
