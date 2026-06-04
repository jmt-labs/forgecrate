## Codegraph-Flavor

Dieses Repo nutzt **codegraph** — einen semantischen Code-Wissensgraphen als MCP-Server.

### Was codegraph bietet

Der MCP-Server läuft lokal (`codegraph serve --mcp`) und stellt folgende Tools bereit:

| Tool | Zweck |
|---|---|
| `codegraph_search` | Semantische Code-Suche ohne exakte Schlüsselwörter |
| `codegraph_node` | Definition eines Symbols (Funktion, Typ, Variable) abrufen |
| `codegraph_callers` / `codegraph_callees` | Alle Aufrufer / Aufgerufenen eines Symbols |
| `codegraph_explore` | Abhängigkeiten und Nachbarn eines Symbols erkunden |
| `codegraph_impact` | Blast-Radius einer Änderung ermitteln |
| `codegraph_files` | Dateien im Index auflisten |
| `codegraph_status` | Index-Status prüfen |

### Pflicht-Regeln

- **Vor jeder nicht-trivialen Änderung MUSS** `codegraph_node` + `codegraph_callers` für betroffene Symbole aufgerufen werden — kein Edit/Write ohne vorherige Codegraph-Abfrage
- **Beim Debuggen MUSS** `codegraph_explore` die Aufrufkette aufzeigen, bevor ein Fix versucht wird
- **Bei Refactoring MUSS** `codegraph_callers` für Call-Sites + `codegraph_search` für Type-/Import-Referenzen geprüft werden
- **Code-Suche**: `codegraph_search` statt grep — grep ist nur erlaubt, wenn codegraph das Ergebnis nicht liefert
- **Impact-Analyse MUSS** `codegraph_impact` vor größeren Umbauten ausgeführt werden

### Index-Aktualisierung

Der Index wird automatisch bei Session-Start im Hintergrund aktualisiert (einmal pro Commit-Stand).
Manuell: `codegraph index` im Repo-Root. Erstmalige Initialisierung: `codegraph init -i` (`--index`): legt `.codegraph/` an und baut den Index headless (keine Prompts).

### Voraussetzung

Installation (einmalig, kein Node.js erforderlich):

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex

# Alternativ via npm
npm i -g @colbymchenry/codegraph
```

Danach im Repo initialisieren (headless, keine Prompts):

```bash
codegraph init -i
```

Der MCP-Server wird über `.mcp.json` automatisch konfiguriert.
