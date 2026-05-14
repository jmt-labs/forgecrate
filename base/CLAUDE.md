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

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer minimal und zielgerichtet
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->
