# Extensions: Deklarative Plugin- & MCP-Installation — Design

**Datum:** 2026-05-14

## Ziel

`claude-setup init` und `claude-setup update` installieren automatisch externe Abhängigkeiten (Claude-Code-Plugins, MCP-Server), die in den Layern deklariert sind. Die Deklaration erfolgt pro Layer in einer `extensions.yaml`. Das Binary merged alle aktiven Layer und ruft die `claude`-CLI auf.

## Dateiformat

Jeder Layer kann eine optionale `extensions.yaml` enthalten:

```yaml
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers

mcp:
  - name: github
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: "${GITHUB_TOKEN}"
```

**Felder Plugins:**
- `name` — eindeutiger Bezeichner (Dedup-Key)
- `source` — Argument für `claude plugin install`

**Felder MCP:**
- `name` — eindeutiger Bezeichner (Dedup-Key)
- `scope` — `local` (Projekt) oder `global` (User); Default: `local`
- `command` — ausführbares Programm
- `args` — optionale Argumente als Liste
- `env` — optionale Umgebungsvariablen als Map

## Merge-Logik

Layer werden in Reihenfolge base → profile → flavors eingelesen. Bei gleichem `name` gewinnt der erste Eintrag (base hat Vorrang). Das Ergebnis ist eine deduplizierte Gesamtliste.

## Installationsschritt

`deploy.Run()` ruft nach `compose.Run()` und `copyHooks()` den neuen Schritt `extensions.Install()` auf.

**Plugins:**
```
claude plugin install <source>
```

**MCP:**
```
claude mcp add <name> --scope <scope> <command> [args...]
```

Umgebungsvariablen aus `env` werden dem Prozess übergeben. Exit-Code ≠ 0 wird mit einer Warnung geloggt, bricht den Deploy aber nicht ab (idempotent-tolerant).

## Neues Package `internal/extensions/`

**`extensions.go`**
- Typen: `Extension`, `Plugin`, `MCP`, `Extensions`
- `Load(path string) (Extensions, error)` — liest eine `extensions.yaml`
- `Merge(layers []Extensions) Extensions` — dedupliziert nach Name, first-wins

**`install.go`**
- `Install(ext Extensions) error` — ruft `claude`-CLI für jeden Eintrag auf

## Betroffene Layer-Dateien

| Datei | Inhalt |
|---|---|
| `base/extensions.yaml` | superpowers Plugin |
| `flavors/devops/extensions.yaml` | (später) Kubernetes MCP etc. |

## Tests

- `Merge` mit überlappenden Namen → first-wins korrekt
- `Merge` mit leeren Listen → kein Fehler
- `Install` shellt korrekt aus (mit Test-Double für `claude`-Binary)
