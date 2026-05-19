# Flavor- und Profil-spezifische Skills — Design

**Datum:** 2026-05-15

## Ziel

Fünf neue SKILL.md-Dateien für bestehende Profile und Flavors, ein neuer `github`-Flavor, und ein Bug-Fix im `forgecrate-advisor`. Kein Go-Code — alles Markdown-Inhaltserstellung. Ein PR.

## Neuer Flavor: `github`

`flavors/github/` wird angelegt mit:

- `CLAUDE.md` — kurze Beschreibung: GitHub-spezifische Workflows (Releases, PR-Templates)
- `extensions.yaml` — leer (kein Plugin erforderlich)
- `skills/github-release/SKILL.md` — Skill (siehe unten)

Der Flavor ist unabhängig vom Profil aktivierbar: `forgecrate run --profile backend --flavor github`.

## Neue Skills

### `flavors/github/skills/github-release/SKILL.md`

Wird **nach** dem base-`release`-Skill aufgerufen. Erstellt ein GitHub Release:

1. Letzten Tag ermitteln: `git describe --tags --abbrev=0`
2. Changelog-Eintrag für diesen Tag extrahieren (Abschnitt zwischen aktuellem und vorherigem Tag-Header)
3. Release erstellen:
   ```bash
   gh release create vX.Y.Z --title "vX.Y.Z" --notes "<extrahierter Changelog>"
   ```
4. Falls Assets vorhanden (erkennbar an `dist/`, `build/`, `*.tar.gz`): per `--attach` hinzufügen
5. Release-URL ausgeben

### `flavors/tdd/skills/test-coverage/SKILL.md`

Analysiert Testabdeckung und schlägt den nächsten Test vor:

1. Coverage-Report erzeugen (sprachspezifisch: `go test -coverprofile=...`, `pytest --cov`, `jest --coverage`)
2. Funktionen/Module mit 0 % Coverage auflisten
3. Für die oberste Lücke: konkreten Testfall vorschlagen (Dateiname, Testname, Eingabe/Erwartung)
4. Fragen: "Soll ich diesen Test anlegen?"

### `flavors/strict-review/skills/pr-checklist/SKILL.md`

Systematische Checkliste vor `gh pr create`:

1. Breaking Changes identifizieren (geänderte Signaturen, entfernte Felder, DB-Migrationen)
2. Testabdeckung für geänderte Dateien prüfen
3. Dokumentation aktuell? (README, CLAUDE.md, Inline-Kommentare)
4. PR-Beschreibung vollständig? (Was, Warum, Wie getestet)
5. Checkliste ausgeben — offene Punkte müssen vor `gh pr create` behoben sein

### `profiles/frontend/skills/accessibility-audit/SKILL.md`

Prüft Barrierefreiheit in geänderten Komponenten:

1. Geänderte Dateien ermitteln (`git diff --name-only`)
2. Für jede UI-Datei prüfen:
   - `<img>` ohne `alt`-Attribut
   - Interaktive Elemente ohne `aria-label` oder sichtbaren Text
   - Formular-Inputs ohne zugehöriges `<label>`
   - Farbkontrast: Warnung wenn Inline-Styles mit hellen Farben auf weißem Hintergrund
3. Befunde mit Datei- und Zeilenverweis ausgeben
4. Bei keinen Befunden: "Keine Barrierefreiheitsprobleme gefunden"

### `profiles/backend/skills/db-migration/SKILL.md`

Führt durch Erstellung und Review einer Datenbankmigrierung:

1. Migrations-Framework erkennen (`golang-migrate`, `flyway`, `alembic`, `prisma`)
2. Neue Migrationsdatei anlegen (korrekter Name mit Timestamp/Sequenznummer)
3. Review-Checkliste:
   - Nicht-destruktiv? (kein `DROP` ohne Datensicherung, kein `NOT NULL` ohne Default)
   - Rollbackfähig? (`DOWN`-Migration vorhanden und korrekt)
   - Performance: Indizes für neue Foreign Keys?
   - Kompatibel mit laufender Anwendung (Blue-Green-fähig)?
4. Checkliste ausgeben — offene Punkte müssen vor dem Commit adressiert sein

## Bug-Fix: `forgecrate-advisor`

`minimal`-Flavor steht in der Tabelle, wird aber in der Entscheidungslogik nie empfohlen. Fix: Schritt 4 bekommt eine dritte Frage:

> "Ist das ein Prototyp oder Solo-Projekt ohne formalen Review-Prozess?"

Wenn ja → `minimal` empfehlen (schließt `strict-review` und `tdd` aus).

## Nicht in Scope

- Skills für `fullstack`-Profil — erbt Frontend- und Backend-Skills automatisch wenn beide Profile aktiv wären; fullstack ist ein eigenständiges Profil, Skills folgen bei Bedarf
- Weitere Flavors (z.B. `docker`, `kubernetes`) — separates Thema
- Automatisches Testen der Skill-Inhalte — wird durch echten Einsatz validiert
