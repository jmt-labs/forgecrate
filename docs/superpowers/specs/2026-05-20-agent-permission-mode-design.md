# Design: Agent Permission Mode

**Datum:** 2026-05-20  
**Status:** Approved

## Ziel

Nutzer können den Claude Code Agent-Berechtigungsmodus (`bypass`, `plan`, `ask`, `auto`) über forgecrate konfigurieren. Der Modus wird in `.forgecrate.yaml` gespeichert und bei Deploy in `settings.json` geschrieben. Ein neuer Subcommand `set-permission-mode` erlaubt nachträgliche Änderungen.

## Datenmodell

`internal/config/config.go` — neues optionales Feld:

```go
type Config struct {
    Version        string            `yaml:"version"`
    Source         string            `yaml:"source"`
    Ref            string            `yaml:"ref"`
    Profile        string            `yaml:"profile"`
    Flavors        []string          `yaml:"flavors"`
    PermissionMode string            `yaml:"permission_mode,omitempty"`
    DeployedFiles  map[string]string `yaml:"deployed_files,omitempty"`
}
```

Gültige Werte: `bypass`, `plan`, `ask`, `auto`. Leer = kein Eintrag in `settings.json`.

`.forgecrate.yaml` nach `forgecrate init`:

```yaml
permission_mode: bypass
```

`settings.json` nach Deploy:

```json
{
  "permissionMode": "bypass",
  "permissions": {
    "allow": ["..."],
    "deny": ["..."]
  }
}
```

`allow`/`deny`-Listen bleiben unverändert neben `permissionMode` (Hard-Rules gelten unabhängig vom Modus).

## Deploy- & Compose-Logik

- `internal/compose/jsonmerge.go` schreibt `permissionMode` als Top-Level-Key in das zusammengesetzte `settings.json`, direkt vor dem `permissions`-Block.
- Bei leerem `PermissionMode` wird der Key weggelassen.
- `settings.json` bleibt eine verwaltete Datei — Hash-Mechanismus greift wie bisher; Konflikte werden normal behandelt.
- Kein neuer Layer, kein Template — der Wert kommt ausschließlich aus `.forgecrate.yaml` via Config.

## Subcommand `forgecrate set-permission-mode <mode>`

Neues File: `cmd/forgecrate/set_permission_mode.go`

Ablauf:
1. `.forgecrate.yaml` lesen
2. `PermissionMode` validieren gegen `{bypass, plan, ask, auto}`
3. `.forgecrate.yaml` zurückschreiben
4. `settings.json` neu generieren und deployen (nur `settings.json`, nicht alle Dateien)

Ausgabe bei Erfolg:
```
✓ permission_mode: bypass
✓ .claude/settings.json aktualisiert
```

Ausgabe bei ungültigem Wert:
```
error: ungültiger Modus "foo" — erlaubt: bypass, plan, ask, auto
```

## `forgecrate init` — Interaktive Abfrage

Neue Frage während `forgecrate init`:

```
Agent permission mode [bypass/plan/ask/auto] (default: bypass):
```

- Enter ohne Eingabe → `bypass`
- Ungültige Eingabe → erneut fragen
- Bestehende Nutzer, die `forgecrate update` ausführen, bekommen `permission_mode` **nicht** automatisch in `.forgecrate.yaml` eingefügt — nur via `set-permission-mode` oder erneutem `init`.

## Fehlerbehandlung

- Ungültiger Modus in `.forgecrate.yaml` → `forgecrate update` bricht mit Fehlermeldung ab
- Fehlendes `.forgecrate.yaml` bei `set-permission-mode` → Fehlermeldung mit Hinweis auf `forgecrate init`

## Tests

| Test | Datei | Art |
|---|---|---|
| Lesen/Schreiben von `permission_mode` | `internal/config/config_test.go` | Unit |
| `permissionMode` korrekt in `settings.json` | `internal/compose/jsonmerge_test.go` | Unit |
| Validierung, YAML-Update, Settings-Deploy | `cmd/forgecrate/set_permission_mode_test.go` | Unit |
| `forgecrate init` + `set-permission-mode` Roundtrip | `e2e/e2e_test.go` | E2E |
