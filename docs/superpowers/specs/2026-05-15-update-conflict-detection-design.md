# Update: Konflikterkennung — Design

**Datum:** 2026-05-15

## Ziel

`claude-setup update` zeigt Konflikte an, wenn der Nutzer eine Datei seit dem letzten Deploy geändert hat und die neue Remote-Version davon abweicht. Basis: Hash-Tracking in `.claude-setup.yaml`.

## Hintergrund

`init` und `update` holen bereits immer frisch von remote (kein Caching). Was fehlt: zu erkennen, ob der Nutzer eine verwaltete Datei manuell geändert hat, bevor sie überschrieben wird.

## Hash-Tracking

### Config-Format (erweitert)

```yaml
version: "1.0"
source: github.com/markus/claude-setup
ref: main
profile: backend
flavors: [tdd]
deployed_files:
  .claude/settings.json: sha256:abc123…
  .claude/hooks/prompt-submit.sh: sha256:def456…
  .claude/hooks/pre-tool.sh: sha256:789abc…
```

`deployed_files` enthält SHA-256-Hashes aller Dateien, die beim letzten Deploy geschrieben wurden (relativ zum Zielverzeichnis).

### Welche Dateien werden getracked

- `.claude/settings.json`
- `.claude/hooks/*.sh`

Nicht getracked:
- `CLAUDE.md` — CUSTOM-Blöcke werden bereits separat bewahrt
- `.claude/skills/*` — first-wins: werden nie überschrieben
- `.claude-setup.yaml` selbst

## Drei-Wege-Vergleich

Für jede zu schreibende Datei beim `update`:

```
hash_disk    = SHA256(aktuelle Datei auf Disk)
hash_stored  = deployed_files[datei]  (aus .claude-setup.yaml)
hash_new     = SHA256(neue Datei von remote)
```

| hash_disk == hash_stored | hash_new == hash_disk | Aktion |
|---|---|---|
| ja | ja | nichts tun (unverändert) |
| ja | nein | einfach überschreiben |
| nein | ja | nichts tun (Nutzer hat geändert, neue Version identisch) |
| nein | nein | **Konflikt** — anzeigen + fragen |

Wenn `hash_stored` fehlt (erste `update` nach Migration): `hash_disk` als "unverändert" behandeln → einfach überschreiben.

## Konfliktausgabe

```
KONFLIKT: .claude/settings.json
  Deine Version: {"model":"claude-opus-4-7","hooks":{}}
  Neue Version:  {"model":"claude-sonnet-4-6","hooks":{}}
  [ü]berschreiben / [b]ehalten (Standard: behalten):
```

- Eingabe `ü` oder `u` → überschreiben
- Alles andere (Enter, `b`) → behalten
- Nach der Entscheidung wird `deployed_files` entsprechend aktualisiert

## Änderungen

### `internal/config/config.go`

`Config`-Struct bekommt ein neues Feld:

```go
DeployedFiles map[string]string `yaml:"deployed_files,omitempty"`
```

### `internal/deploy/deploy.go`

Neue Funktion `deployFile(dst, rel, newContent string, cfg *config.Config, w io.Writer) error`:

1. Hash von neuem Inhalt berechnen
2. Drei-Wege-Vergleich durchführen
3. Bei Konflikt: Ausgabe auf `w`, Eingabe lesen, entscheiden
4. Datei schreiben oder überspringen
5. `cfg.DeployedFiles[rel]` aktualisieren

`copyHooks` und der Settings-Kopierschritt nutzen `deployFile` statt direktem `os.WriteFile`.

`w io.Writer` und eine `reader io.Reader`-Abhängigkeit ermöglichen Tests ohne interaktive Eingabe.

### `cmd/claude-setup/update.go`

Nach `deploy.Run(...)`: `config.Write(cfgPath, cfg)` bereits vorhanden — schreibt die aktualisierten Hashes zurück.

### `cmd/claude-setup/init.go`

Nach dem ersten Deploy ebenfalls `deployed_files` befüllen, damit `update` beim nächsten Aufruf einen Baseline-Hash hat.

## Tests

- Unit-Test in `internal/deploy/`: Konflikt wird erkannt und gemeldet (Eingabe simuliert)
- Unit-Test: Kein Konflikt wenn Nutzer nichts geändert hat
- Unit-Test: Kein Konflikt wenn neue Version == aktuelle Version (Nutzer hat gleich geändert)
- Unit-Test: Fehlender `hash_stored` → kein Konflikt (Migration)

## Nicht in Scope

- Merge-Strategie (automatisches Zusammenführen von JSON-Änderungen)
- `--force`-Flag zum Überspringen aller Konflikte — kann später ergänzt werden
- Konflikterkennung für `CLAUDE.md` — CUSTOM-Blöcke decken den relevanten Fall ab
