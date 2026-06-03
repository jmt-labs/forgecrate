## GETBETTER-Flavor

Kontinuierliche Verbesserung durch Festhalten von Erkenntnissen aus jeder Session.

- Am Session-Start: `mcp__memory__read_graph` aufrufen, Entities vom Typ `session-reflection` lesen.
- Am Sessionende: `/forgecrate-getbetter` aufrufen um Erkenntnisse zu speichern.

**Was gespeichert wird (memory MCP, Entity `session-reflection`):**
- Wiederkehrende Fehler und deren Ursachen
- Patterns die gut funktioniert haben
- Entscheidungen die sich im Nachhinein als falsch erwiesen haben
- Projektspezifische Gotchas die nicht aus dem Code ersichtlich sind

**Format pro Erkenntnis:** `[YYYY-MM-DD] <Kategorie>: <Erkenntnis in einem Satz>`
Kategorien: `workflow`, `tooling`, `pattern`, `mistake`, `decision`.
