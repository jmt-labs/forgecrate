## Codegraph-Flavor

Dieses Repo nutzt **codegraph** — einen semantischen Code-Wissensgraphen als MCP-Server.

### Was codegraph bietet

Der MCP-Server läuft lokal (`codegraph serve --mcp`) und stellt 7 Tools bereit:

| Tool | Zweck |
|---|---|
| `search_code` | Semantische Code-Suche ohne exakte Schlüsselwörter |
| `get_definition` | Definition eines Symbols (Funktion, Typ, Variable) abrufen |
| `find_references` | Alle Verwendungen eines Symbols im Repo finden |
| `get_call_graph` | Aufrufgraph für eine Funktion erstellen |
| `get_dependencies` | Abhängigkeiten eines Moduls/Pakets auflisten |
| `explain_code` | Code-Abschnitt mit Graph-Kontext erklären |
| `find_similar` | Ähnliche Code-Muster im gesamten Repo finden |

### Wann nutzen

- **Vor jeder nicht-trivialen Änderung**: `get_definition` + `find_references` für betroffene Symbole
- **Beim Debuggen**: `get_call_graph` um Aufrufkette nachzuvollziehen
- **Bei Refactoring**: `find_references` sicherstellt vollständige Erfassung aller Verwendungen
- **Code-Suche**: `search_code` statt grep bei konzeptuellen Fragen

### Index-Aktualisierung

Der Index wird automatisch nach jedem Commit aktualisiert (post-commit Hook).
Manuell: `codegraph index .` im Repo-Root.

### Voraussetzung

codegraph muss installiert sein: `pip install codegraph` oder `go install github.com/schollz/codegraph@latest`.
Der MCP-Server wird über `.mcp.json` automatisch konfiguriert.
