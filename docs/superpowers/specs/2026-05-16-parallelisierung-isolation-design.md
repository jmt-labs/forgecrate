# Design: Parallelisierung & Isolation

## Ziel

Claude soll eigenständig entscheiden, wann Subagenten im Hintergrund
dispatcht werden (`run_in_background: true`) und wann sie in einem
isolierten Git-Worktree laufen (`isolation: "worktree"`) — ohne
explizite Aufforderung durch den Nutzer.

## Hintergrund

Aktuell nutzt Claude Background-Dispatch und Worktree-Isolation nur
wenn explizit darum gebeten wird. Die Entscheidungslogik fehlt im
Base-Layer. Superpowers-Skills für beide Mechanismen existieren bereits
(`superpowers:dispatching-parallel-agents`, `superpowers:using-git-worktrees`),
sind aber nicht als Default-Verhalten verankert.

## Design

### Ansatz

Kompakte Entscheidungsmatrix direkt in `base/CLAUDE.md` — unmittelbar
im Kontext, kein Skill-Lookup nötig für die Entscheidung selbst.
Die Superpowers-Skills werden als Referenz für den konkreten Ablauf
verlinkt.

### Neuer Abschnitt in `base/CLAUDE.md`

Position: nach "Team-Rollen & Subagent-Konfiguration", vor "MCP Server".

```markdown
## Parallelisierung & Isolation

Subagenten werden proaktiv parallelisiert und isoliert — ohne explizite Aufforderung.

| Situation | Mechanismus | Anleitung |
|---|---|---|
| Task dauert >1 min oder Ergebnis nicht sofort nötig | `run_in_background: true` | `superpowers:dispatching-parallel-agents` |
| Feature-Branch, Multi-File-Änderung, langer Plan | `isolation: "worktree"` | `superpowers:using-git-worktrees` |
| Mehrere unabhängige Tasks gleichzeitig | beide kombinieren | beide Skills |

Im Zweifelsfall Background nutzen — warten ist kein Default.
```

### Entscheidungslogik

**Background-Dispatch** (`run_in_background: true`):
- Geschätzte Laufzeit >1 Minute (Builds, Tests, Research, API-Calls)
- Ergebnis wird nicht sofort im nächsten Schritt benötigt
- Mehrere voneinander unabhängige Tasks können parallel laufen

**Worktree-Isolation** (`isolation: "worktree"`):
- Implementierung auf einem Feature-Branch
- Änderungen betreffen viele Dateien gleichzeitig
- Langer Implementierungsplan (>1 Task)
- Hauptworkspace soll während der Arbeit sauber bleiben

**Keines von beidem:**
- Kurze Tasks (<1 Minute, z.B. einzelne Datei lesen/schreiben)
- Task hängt direkt vom aktuellen Workspace-State ab
- Sequentielle Abhängigkeit zum nächsten Schritt

## Scope

- **Modify:** `base/CLAUDE.md` — neuer Abschnitt einfügen
- **Kein Go-Code**, keine neuen Dateien, keine neuen Skills

## Testbarkeit

E2E-Test: Nach Deploy enthält `CLAUDE.md` im Zielprojekt den
"Parallelisierung & Isolation"-Abschnitt mit der Entscheidungsmatrix.
