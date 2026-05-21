# Design: `forgecrate config` — Interaktiver Konfigurationsbefehl

**Datum:** 2026-05-21  
**Status:** Approved

## Ziel

Nutzer sollen Profil und Flavors eines bestehenden forgecrate-Repos nachträglich ändern können — ohne manuell YAML zu editieren und ohne `forgecrate init` erneut aufzurufen. Ein einzelner Befehl öffnet einen interaktiven Wizard, der die Auswahl mit Pfeilnavigation ermöglicht und danach sofort deployed.

## Kommando

```
forgecrate config
```

Neuer Top-Level-Command, registriert neben `init`, `update`, `list`, `describe`, `hook`.

## Ablauf

1. `.forgecrate.yaml` im CWD lesen — Fehler wenn nicht vorhanden: *„Kein forgecrate-Repo. Zuerst `forgecrate init` ausführen."*
2. Source-Repo in ein temporäres Verzeichnis klonen (wie `describe.go` via `gh.Clone`)
3. Verfügbare Profile aus `profiles/` lesen (`os.ReadDir`)
4. Verfügbare Flavors aus `flavors/` lesen (`os.ReadDir`)
5. Interaktives Formular (`charmbracelet/huh`):
   - `huh.NewSelect` — Profil wählen, aktueller Wert vorausgewählt
   - `huh.NewMultiSelect` — Flavors togglen (Leertaste), aktuelle Werte vorausgewählt
6. `.forgecrate.yaml` mit neuen Werten schreiben (`config.Write`)
7. `deploy.Run(sourceDir, cwd, newCfg)` — volles Deploy (identisch zu `forgecrate update`)

## UX-Skizze

```
$ forgecrate config

 Profil
 > backend
   frontend
   fullstack

 Flavors  (Leertaste = toggle)
 > [x] tdd
   [x] strict-review
   [ ] github
   [ ] gitops
   [ ] minimal
   [ ] no-research
   [ ] getbetter

Deploye...
✅ CLAUDE.md
✅ settings.json
✅ .claude/hooks/pre-tool.sh
```

## Architektur

### Neue Dateien

| Datei | Inhalt |
|---|---|
| `cmd/forgecrate/config.go` | `newConfigCmd()` + `configInteractive(dir, in, out)` |
| `cmd/forgecrate/config_test.go` | Unit-Tests ohne echten Upstream |

### Geänderte Dateien

| Datei | Änderung |
|---|---|
| `cmd/forgecrate/main.go` | `root.AddCommand(newConfigCmd())` |
| `go.mod` / `go.sum` | `github.com/charmbracelet/huh` |

### Kernfunktion

```go
func configInteractive(dir string) error {
    // 1. Config lesen
    cfg, err := config.Read(filepath.Join(dir, ".forgecrate.yaml"))
    // 2. Upstream klonen
    srcDir, cleanup, err := cloneSource(cfg)
    defer cleanup()
    // 3. Optionen einlesen
    profiles := listDirs(filepath.Join(srcDir, "profiles"))
    flavors  := listDirs(filepath.Join(srcDir, "flavors"))
    // 4. huh-Formular
    var newProfile string
    var newFlavors []string
    huh.NewForm(...).Run()
    // 5. Schreiben + Deployen
    cfg.Profile = newProfile
    cfg.Flavors = newFlavors
    config.Write(...)
    return deploy.Run(srcDir, dir, cfg)
}
```

### Testabdeckung

Business-Logik und huh-Form werden getrennt: `configInteractive` bekommt eine `promptFn func(profiles, flavors []string, cur config.Config) (string, []string, error)` als Parameter. In Production wird `huhPrompt` übergeben, in Tests eine Stub-Funktion.

- `configInteractive` mit gemocktem Verzeichnis und Stub-`promptFn`
- Fehlerfall: `.forgecrate.yaml` fehlt → klarer Fehler
- Profil-/Flavor-Listing korrekt aus Verzeichnis
- Config wird korrekt geschrieben (YAML-Roundtrip)
- `promptFn`-Fehler wird weitergegeben

## Abhängigkeiten

- `github.com/charmbracelet/huh` — interaktive Prompts mit Pfeilnavigation
- Bestehende interne Packages: `config`, `deploy`, `github`

## Error Handling

| Situation | Verhalten |
|---|---|
| Kein `.forgecrate.yaml` | Fehler: „Kein forgecrate-Repo. Zuerst `forgecrate init` ausführen." |
| Source-Repo nicht erreichbar | Fehler weitergeben (wie `describe`) |
| Keine Profile/Flavors gefunden | Fehler: „Keine Profile/Flavors im Source-Repo gefunden" |
| Deploy-Fehler | Fehler weitergeben, Config wurde bereits gespeichert |
