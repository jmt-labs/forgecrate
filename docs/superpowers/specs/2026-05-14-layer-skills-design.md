# Layer Skills — Design

**Datum:** 2026-05-14
**Issue:** folgt

## Ziel

Vier projekt-unabhängige Skills werden über das Layer-System in Ziel-Repos verteilt: Release-Workflow, Repo-Onboarding, Repo-Health-Analyse und ein claude-setup-Advisor. Alle landen in `.claude/skills/` des Ziel-Repos und stehen dort per `/skill-name` zur Verfügung.

## Delivery-Mechanismus

### Verzeichnisstruktur im claude-setup Repo

```
base/skills/
  release/SKILL.md
  repo-onboarding/SKILL.md
  repo-health/SKILL.md
  claude-setup-advisor/SKILL.md
profiles/frontend/skills/   # leer — Mechanismus steht für spätere Skills bereit
flavors/strict-review/skills/  # leer — dto.
```

### Kopierschritt in `deploy.go`

`RunWithClaude()` bekommt nach `installExtensions()` einen neuen Schritt: `copySkills(sourceDir, destDir, cfg)`.

```
func copySkills(sourceDir, destDir string, cfg config.Config) error
```

Verarbeitungsreihenfolge (base → profile → flavor):

```
base/skills/
profiles/<profile>/skills/
flavors/<flavor>/skills/   (für jeden aktiven Flavor)
```

Für jedes Skill-Verzeichnis (`<name>/`):
- Ziel: `<destDir>/.claude/skills/<name>/`
- **First-wins:** existiert das Ziel bereits, wird es übersprungen — kein Überschreiben
- Alle Dateien im Verzeichnis werden rekursiv kopiert

First-wins nach Layer-Reihenfolge (base → profile → flavor): base gewinnt bei Namenskonflikt. Profile und Flavors können keine base-Skills überschreiben — sie verwenden unterschiedliche Namen. Das ist dasselbe Prinzip wie `extensions.Merge()`.

### Tests

`internal/deploy/deploy_test.go` bekommt Tests für `copySkills()`:
- Skill aus base landet in `.claude/skills/`
- Skill aus Profile landet ebenfalls, wenn kein Namenskonflikt
- Namenskonflikt: base gewinnt (first-wins)
- Fehlende `skills/`-Verzeichnisse werden stillschweigend übersprungen (kein Fehler)

## Skill-Inhalte

Alle vier Skills liegen in `base/skills/` und sind damit in jedem Repo verfügbar, unabhängig von Profil und Flavor.

### `release/SKILL.md`

Führt durch einen vollständigen Release-Zyklus:
1. Test-Suite ausführen — bei Fehlern stoppen
2. Changelog prüfen / ergänzen
3. Version-Tag erstellen (SemVer)
4. Tag pushen
5. CI-Status prüfen (falls GitHub Actions vorhanden)

Der Skill erkennt den eingesetzten Paketmanager und Build-Mechanismus aus dem Repo (z.B. `go.mod`, `package.json`, `Makefile`) und passt die Befehle an.

### `repo-onboarding/SKILL.md`

Erkundet das Repo nach `claude-setup run` und erstellt eine strukturierte Zusammenfassung:
- Welche Sprachen/Frameworks sind im Einsatz
- Wo liegen Tests, Business-Logik, Konfiguration
- Wie wird gebaut, wie wird getestet
- Welche externen Abhängigkeiten gibt es

Ausgabe: Vorschlag für den GENERATED-Block in `CLAUDE.md`. Nutzer kann übernehmen oder anpassen.

### `repo-health/SKILL.md`

Analysiert das Repo auf Verbesserungspotenzial und gibt eine priorisierte Liste zurück:
- Testabdeckung (fehlende Tests für öffentliche Funktionen)
- Veraltete Abhängigkeiten
- Ungenutzte Exports / Dead Code
- Fehlende oder veraltete Dokumentation
- Bekannte Sicherheitsmuster (hardcoded secrets, unsichere Patterns)

Ausgabe: Nummerierte Liste nach Priorität, jeweils mit konkretem Datei-/Zeilenverweis.

### `claude-setup-advisor/SKILL.md`

Analysiert ein Repo und empfiehlt das passende claude-setup-Profil und -Flavor:
1. Sprache und Framework erkennen → Profil (backend/frontend/fullstack)
2. Test-Konventionen erkennen → Flavor `tdd` sinnvoll?
3. Review-Anforderungen abfragen → Flavor `strict-review` sinnvoll?
4. Gibt den exakten `claude-setup run`-Befehl aus:
   ```
   claude-setup run --profile frontend --flavor strict-review
   ```

Der Skill kennt alle verfügbaren Profile und Flavors und erklärt kurz, warum er die jeweilige Kombination empfiehlt.

## Nicht in Scope

- Profil- oder flavor-spezifische Skills in dieser Iteration — Mechanismus steht, Inhalte folgen bei Bedarf
- Automatisches Aktualisieren bestehender Skills beim Re-Run von `claude-setup` — first-wins gilt auch für Re-Runs
- Skills für das claude-setup Repo selbst (Entwickler-Skills) — separates Thema
