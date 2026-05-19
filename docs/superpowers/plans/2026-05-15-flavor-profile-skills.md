# Flavor- und Profil-spezifische Skills — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fünf neue SKILL.md-Dateien für Flavors und Profile, einen neuen `github`-Flavor und einen Bug-Fix im `forgecrate-advisor` — alles Markdown-Inhalt, kein Go-Code außer E2E-Tests.

**Architecture:** Alle Skill-Dateien landen in `flavors/<name>/skills/<skill>/SKILL.md` bzw. `profiles/<name>/skills/<skill>/SKILL.md`. Der `copySkills`-Mechanismus in `deploy.go` kopiert sie automatisch ins Zielrepo. E2E-Tests prüfen den vollständigen Deploy-Pfad mit den echten Dateien.

**Tech Stack:** Go (E2E-Tests), Markdown (Skill-Inhalt)

---

### Task 1: E2E-Tests für copySkills (failing zuerst)

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Failing tests schreiben**

Füge am Ende von `e2e/e2e_test.go` vor der letzten `}` ein:

```go
func TestDeployIncludesBaseSkills(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	skills := []string{"release", "repo-onboarding", "repo-health", "forgecrate-advisor"}
	for _, s := range skills {
		path := filepath.Join(dst, ".claude", "skills", s, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("base skill missing: %s", s)
		}
	}
}

func TestDeployIncludesProfileSkill(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "frontend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "skills", "accessibility-audit", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("frontend profile skill missing: accessibility-audit")
	}
}

func TestDeployIncludesFlavorSkill(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"github"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "skills", "github-release", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("github flavor skill missing: github-release")
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```bash
go test ./e2e/... -run "TestDeployIncludesProfileSkill|TestDeployIncludesFlavorSkill" -v
```

Erwartung: FAIL — Dateien existieren noch nicht.

`TestDeployIncludesBaseSkills` sollte bereits bestehen (base skills sind vorhanden).

- [ ] **Step 3: Commit**

```bash
git add e2e/e2e_test.go
git commit -m "test(e2e): add copySkills E2E tests — red phase"
```

---

### Task 2: `github`-Flavor anlegen

**Files:**
- Create: `flavors/github/CLAUDE.md`
- Create: `flavors/github/extensions.yaml`
- Create: `flavors/github/skills/github-release/SKILL.md`

- [ ] **Step 1: `flavors/github/CLAUDE.md` erstellen**

```markdown
## GitHub-Flavor

- Releases über `gh release create` veröffentlichen (nach `release`-Skill)
- PR-Templates in `.github/pull_request_template.md` pflegen
- CI-Status mit `gh run list` prüfen bevor ein Release getaggt wird
```

- [ ] **Step 2: `flavors/github/extensions.yaml` erstellen**

```yaml
plugins: []
mcpServers: []
```

- [ ] **Step 3: `flavors/github/skills/github-release/SKILL.md` erstellen**

```markdown
# GitHub Release

Wird **nach** dem `release`-Skill aufgerufen. Erstellt ein GitHub Release für den soeben erstellten Tag.

## Voraussetzung

`gh` CLI ist installiert und authentifiziert (`gh auth status`).

## Ablauf

1. **Letzten Tag ermitteln**
   ```bash
   git describe --tags --abbrev=0
   ```

2. **Changelog-Abschnitt extrahieren** — suche in `CHANGELOG.md` den Abschnitt zwischen dem aktuellen und dem vorherigen Tag-Header (Format `## vX.Y.Z` oder `## [vX.Y.Z]`). Falls kein CHANGELOG vorhanden: frage nach Release Notes.

3. **GitHub Release erstellen**
   ```bash
   gh release create vX.Y.Z \
     --title "vX.Y.Z" \
     --notes "<extrahierter Changelog-Abschnitt>"
   ```

4. **Assets anhängen** — prüfe ob `dist/`, `build/` oder `*.tar.gz`-Dateien vorhanden sind:
   ```bash
   gh release upload vX.Y.Z dist/* 2>/dev/null || true
   ```

5. **Release-URL ausgeben**
   ```bash
   gh release view vX.Y.Z --json url -q .url
   ```

## Hinweise

- Läuft erst nach erfolgreichem `release`-Skill (Tag bereits gepusht).
- Bei Fehler (`gh` nicht installiert, nicht authentifiziert): Fehlermeldung ausgeben, Schritt überspringen.
```

- [ ] **Step 4: Tests ausführen — `TestDeployIncludesFlavorSkill` muss bestehen**

```bash
go test ./e2e/... -run "TestDeployIncludesFlavorSkill" -v
```

Erwartung: PASS

- [ ] **Step 5: Commit**

```bash
git add flavors/github/
git commit -m "feat(github): add github flavor with github-release skill"
```

---

### Task 3: `tdd`-Flavor — `test-coverage`-Skill

**Files:**
- Create: `flavors/tdd/skills/test-coverage/SKILL.md`

- [ ] **Step 1: Datei erstellen**

```markdown
# Test Coverage

Analysiert Testabdeckung und schlägt den nächsten konkreten Test vor.

## Ablauf

1. **Coverage-Report erzeugen** — erkenne Build-System:
   - Go: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`
   - Node/Jest: `npm test -- --coverage 2>/dev/null || npx jest --coverage`
   - Python: `pytest --cov --cov-report=term-missing 2>/dev/null`

2. **Lücken auflisten** — filtere Zeilen mit `0.0%` Coverage und zeige die Top 5:
   ```
   internal/deploy/deploy.go:copyFile     0.0%
   internal/config/config.go:Validate     0.0%
   ...
   ```

3. **Nächsten Test vorschlagen** — für die oberste Lücke (höchste Zeilenzahl ohne Coverage):
   - Funktion/Methode benennen
   - Sinnvollen Testfall formulieren: Eingabe, Erwartung, Randfall
   - Dateiname und Testfunktionsname vorschlagen

   Beispiel:
   ```
   Vorschlag: internal/deploy/deploy_test.go
   func TestCopyFileCreatesParentDirs(t *testing.T) {
     // prüft ob copyFile fehlende Parent-Verzeichnisse anlegt
   }
   ```

4. **Fragen** — "Soll ich diesen Test anlegen?"

## Hinweise

- Läuft nach jedem Feature-Zyklus, nicht nur bei roten Tests.
- Fokus auf Verhalten, nicht auf Line-Coverage als Selbstzweck.
```

- [ ] **Step 2: Tests ausführen**

```bash
go test ./... -v 2>&1 | tail -20
```

Erwartung: alle Tests PASS (keine Go-Änderungen).

- [ ] **Step 3: Commit**

```bash
git add flavors/tdd/skills/test-coverage/SKILL.md
git commit -m "feat(tdd): add test-coverage skill"
```

---

### Task 4: `strict-review`-Flavor — `pr-checklist`-Skill

**Files:**
- Create: `flavors/strict-review/skills/pr-checklist/SKILL.md`

- [ ] **Step 1: Datei erstellen**

```markdown
# PR-Checkliste

Systematische Überprüfung vor `gh pr create`. Alle offenen Punkte müssen adressiert sein.

## Ablauf

1. **Breaking Changes identifizieren**
   - Geänderte Funktionssignaturen: `git diff main -- '*.go' '*.ts' '*.py'` nach `func ` / `export` suchen
   - Entfernte oder umbenannte Felder in structs/types
   - Datenbankmigrationen vorhanden?

2. **Testabdeckung für geänderte Dateien**
   ```bash
   git diff main --name-only | grep -v '_test\.' | while read f; do
     base="${f%.*}"
     test_candidates=("${base}_test.go" "${base}.test.ts" "tests/test_${base##*/}.py")
     for t in "${test_candidates[@]}"; do [ -f "$t" ] && echo "OK: $t" && break; done
   done
   ```
   Dateien ohne zugehörigen Test: explizit auflisten.

3. **Dokumentation aktuell?**
   - `README.md` — deckt neue Features/Flags ab?
   - `CLAUDE.md` — GENERATED-Block aktuell?
   - Inline-Kommentare für nicht-offensichtliches Verhalten?

4. **PR-Beschreibung vollständig?** Prüfe ob folgende Punkte beantwortet sind:
   - Was wurde geändert?
   - Warum (Ticket/Issue verlinkt)?
   - Wie wurde getestet?

5. **Checkliste ausgeben**

   ```
   ✅ Keine Breaking Changes erkannt
   ⚠️  internal/deploy/deploy.go hat keinen zugehörigen Test für copyFile
   ✅ README.md aktuell
   ❌ PR-Beschreibung: "Wie getestet" fehlt
   ```

   Offene ❌-Punkte müssen behoben sein bevor `gh pr create` aufgerufen wird.
```

- [ ] **Step 2: Tests ausführen**

```bash
go test ./... -v 2>&1 | tail -10
```

Erwartung: PASS

- [ ] **Step 3: Commit**

```bash
git add flavors/strict-review/skills/pr-checklist/SKILL.md
git commit -m "feat(strict-review): add pr-checklist skill"
```

---

### Task 5: `frontend`-Profil — `accessibility-audit`-Skill

**Files:**
- Create: `profiles/frontend/skills/accessibility-audit/SKILL.md`

- [ ] **Step 1: Datei erstellen**

```markdown
# Accessibility Audit

Prüft Barrierefreiheit in geänderten UI-Komponenten.

## Ablauf

1. **Geänderte Dateien ermitteln**
   ```bash
   git diff --name-only | grep -E '\.(tsx?|jsx?|vue|svelte|html)$'
   ```

2. **Für jede Datei prüfen:**

   **Fehlende `alt`-Attribute:**
   ```bash
   grep -n '<img' <datei> | grep -v 'alt='
   ```

   **Interaktive Elemente ohne Label:**
   ```bash
   grep -n '<button\|<a ' <datei> | grep -v 'aria-label\|aria-labelledby'
   ```
   Prüfe zusätzlich ob der Button/Link sichtbaren Textinhalt hat (kein reines Icon ohne Label).

   **Formular-Inputs ohne `<label>`:**
   ```bash
   grep -n '<input\|<textarea\|<select' <datei> | grep -v 'type="hidden"'
   ```
   Für jeden Fund: prüfe ob ein `<label for="...">` mit passendem `id` existiert.

   **Inline-Farbkontrast-Warnung:**
   ```bash
   grep -n 'color:.*#[fF][fF]\|color:.*white\|color:.*rgb(255' <datei>
   ```
   Hinweis wenn helle Textfarbe auf weißem Hintergrund vermutet wird.

3. **Ausgabe** — Befunde mit Datei und Zeile:
   ```
   components/Button.tsx:12  <img> ohne alt-Attribut
   components/Nav.tsx:34     <button> ohne aria-label und ohne sichtbaren Text
   ```
   Bei keinen Befunden: "Keine Barrierefreiheitsprobleme in geänderten Dateien gefunden."

## Hinweise

- Kein Ersatz für echte Screenreader-Tests — deckt die häufigsten statischen Fehler ab.
- Bei generierten Dateien (`.min.js`, Build-Output) überspringen.
```

- [ ] **Step 2: Tests ausführen**

```bash
go test ./... -v 2>&1 | tail -10
```

Erwartung: PASS

- [ ] **Step 3: Commit**

```bash
git add profiles/frontend/skills/accessibility-audit/SKILL.md
git commit -m "feat(frontend): add accessibility-audit skill"
```

---

### Task 6: `backend`-Profil — `db-migration`-Skill

**Files:**
- Create: `profiles/backend/skills/db-migration/SKILL.md`

- [ ] **Step 1: Datei erstellen**

```markdown
# DB Migration

Führt durch Erstellung und Review einer Datenbankmigrierung.

## Framework erkennen

Prüfe in dieser Reihenfolge:
- `golang-migrate`: Verzeichnis `migrations/` mit `*.up.sql`/`*.down.sql`
- `flyway`: Verzeichnis `db/migration/` mit `V__*.sql`
- `alembic`: `alembic.ini` + `alembic/versions/`
- `prisma`: `prisma/schema.prisma` + `prisma/migrations/`
- Kein Framework erkannt: frage welches verwendet wird

## Ablauf

1. **Neue Migrationsdatei anlegen**

   *golang-migrate:*
   ```bash
   migrate create -ext sql -dir migrations -seq <name>
   # erzeugt: migrations/000N_<name>.up.sql und .down.sql
   ```

   *flyway:*
   ```bash
   touch db/migration/V$(date +%Y%m%d%H%M%S)__<name>.sql
   ```

   *alembic:*
   ```bash
   alembic revision --autogenerate -m "<name>"
   ```

   *prisma:*
   ```bash
   npx prisma migrate dev --name <name>
   ```

2. **Review-Checkliste**

   Überprüfe die erstellte Migration auf:

   **Nicht-destruktiv?**
   - `DROP TABLE` / `DROP COLUMN`: Datensicherung oder Feature-Flag vorhanden?
   - `NOT NULL`-Spalte hinzufügen: Default-Wert oder zweistufige Migration (Spalte nullable → befüllen → NOT NULL)?

   **Rollbackfähig?**
   - `DOWN`-Migration vorhanden und spiegelt `UP` korrekt?
   - Bei `alembic`/`prisma`: `downgrade`-Funktion implementiert?

   **Performance?**
   - Neue Foreign Keys: Index angelegt?
   - Große Tabellen (>100k Zeilen): `CONCURRENTLY`-Index oder Batch-Update nötig?

   **Blue-Green-kompatibel?**
   - Läuft die Anwendung mit der alten UND der neuen Schema-Version gleichzeitig?
   - Spalten-Umbenennungen: zweistufig (neue Spalte → Daten kopieren → alte Spalte entfernen)?

3. **Ausgabe**
   ```
   ✅ DOWN-Migration vorhanden
   ✅ Kein DROP ohne Sicherung
   ⚠️  Neue Foreign Key-Spalte ohne Index — empfehle: CREATE INDEX CONCURRENTLY
   ❌ NOT NULL ohne Default — bestehende Zeilen werden beim Migrate fehlschlagen
   ```

   Offene ❌-Punkte müssen vor dem Commit behoben sein.
```

- [ ] **Step 2: Tests ausführen**

```bash
go test ./... -v 2>&1 | tail -10
```

Erwartung: PASS

- [ ] **Step 3: E2E-Test bestätigen**

```bash
go test ./e2e/... -run "TestDeployIncludesBaseSkills|TestDeployIncludesProfileSkill|TestDeployIncludesFlavorSkill" -v
```

Erwartung: alle drei PASS (alle Skills vorhanden).

- [ ] **Step 4: Commit**

```bash
git add profiles/backend/skills/db-migration/SKILL.md
git commit -m "feat(backend): add db-migration skill"
```

---

### Task 7: Bug-Fix `forgecrate-advisor` — `minimal`-Flavor

**Files:**
- Modify: `base/skills/forgecrate-advisor/SKILL.md`

- [ ] **Step 1: Datei öffnen und Schritt 4 ersetzen**

Aktuelle Schritt-4-Zeile:

```markdown
4. **Review-Anforderungen abfragen** — stelle eine Frage: "Arbeitest du alleine oder im Team mit PR-Reviews?" → Flavor `strict-review` sinnvoll?
```

Ersetzen durch:

```markdown
4. **Arbeitsweise abfragen** — stelle diese Fragen nacheinander:

   a. "Ist das ein Prototyp oder Solo-Projekt ohne formalen Review-Prozess?"
      - Ja → empfehle `minimal` (schließt `strict-review` und `tdd` aus, weiter mit Schritt 5)
      - Nein → weiter mit b

   b. "Arbeitest du im Team mit PR-Reviews?"
      - Ja → Flavor `strict-review` vormerken

   c. "Schreibst du Tests vor der Implementierung (Test-first)?"
      - Ja → Flavor `tdd` vormerken
```

- [ ] **Step 2: Alle Tests ausführen**

```bash
go test ./...
```

Erwartung: PASS

- [ ] **Step 3: Commit und PR**

```bash
git add base/skills/forgecrate-advisor/SKILL.md
git commit -m "fix(advisor): add minimal flavor to decision logic"
```

```bash
git push -u origin feat/flavor-profile-skills
gh pr create \
  --title "feat: flavor- und profil-spezifische Skills" \
  --body "Closes #10

## Inhalt
- Neuer Flavor \`github\` mit \`github-release\`-Skill
- \`tdd\`: \`test-coverage\`-Skill
- \`strict-review\`: \`pr-checklist\`-Skill
- \`frontend\`: \`accessibility-audit\`-Skill
- \`backend\`: \`db-migration\`-Skill
- Fix: \`minimal\`-Flavor in \`forgecrate-advisor\`
- E2E-Tests für \`copySkills\`

## Test Plan
- [ ] \`TestDeployIncludesBaseSkills\` — base Skills vorhanden
- [ ] \`TestDeployIncludesProfileSkill\` — frontend/accessibility-audit vorhanden
- [ ] \`TestDeployIncludesFlavorSkill\` — github/github-release vorhanden
- [ ] Alle bestehenden Tests grün"
```
