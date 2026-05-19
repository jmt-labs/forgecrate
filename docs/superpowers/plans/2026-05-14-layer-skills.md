# Layer Skills Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Einen `copySkills()`-Schritt in `deploy.go` einbauen, der Skill-Verzeichnisse aus dem Layer-System nach `.claude/skills/` im Ziel-Repo kopiert, und vier SKILL.md-Dateien anlegen.

**Architecture:** Reine Go-Implementierung in `internal/deploy/deploy.go`. Neuer Schritt `copySkills()` nach `installExtensions()`. First-wins-Dedup nach Layer-Reihenfolge (base → profile → flavor). Kein neues Paket, keine neuen Abhängigkeiten.

**Tech Stack:** Go 1.24, `go test ./...`

**Abhängigkeit:** keine — `deploy.go` ist vollständig implementiert.

---

## Dateistruktur

| Datei | Aktion | Zweck |
|---|---|---|
| `internal/deploy/deploy.go` | Ändern | `copySkills()`, `copyDir()`, `copyFile()` hinzufügen; in `RunWithClaude()` verdrahten |
| `internal/deploy/deploy_test.go` | Ändern | 4 neue Tests für `copySkills()` |
| `base/skills/release/SKILL.md` | Neu | Release-Workflow-Skill |
| `base/skills/repo-onboarding/SKILL.md` | Neu | Repo-Überblick-Skill |
| `base/skills/repo-health/SKILL.md` | Neu | Repo-Analyse-Skill |
| `base/skills/forgecrate-advisor/SKILL.md` | Neu | Profil/Flavor-Empfehlungs-Skill |

---

### Task 1: `copySkills()` — TDD

**Files:**
- Modify: `internal/deploy/deploy_test.go`
- Modify: `internal/deploy/deploy.go`

- [ ] **Schritt 1: 4 Failing Tests schreiben**

Füge in `internal/deploy/deploy_test.go` nach `TestRunInstallsExtensions` hinzu:

```go
func TestRunCopiesSkillsFromBase(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "base/skills/release/SKILL.md", "# Release Skill")

	cfg := config.Config{Profile: "backend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "release", "SKILL.md"))
	if err != nil {
		t.Fatalf("skill not copied: %v", err)
	}
	if string(got) != "# Release Skill" {
		t.Errorf("content: %q", got)
	}
}

func TestCopySkillsFirstWins(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "base/skills/release/SKILL.md", "base-content")
	writeFile(t, src, "profiles/frontend/skills/release/SKILL.md", "profile-content")

	cfg := config.Config{Profile: "frontend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(dst, ".claude", "skills", "release", "SKILL.md"))
	if string(got) != "base-content" {
		t.Errorf("first-wins failed: got %q, want %q", string(got), "base-content")
	}
}

func TestCopySkillsMissingDirOK(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)

	cfg := config.Config{Profile: "backend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("no error expected when skills dir missing: %v", err)
	}
}

func TestCopySkillsProfileAndFlavor(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "profiles/frontend/skills/frontend-tips/SKILL.md", "frontend-tips")
	writeFile(t, src, "flavors/strict-review/skills/review-tips/SKILL.md", "review-tips")

	cfg := config.Config{Profile: "frontend", Flavors: []string{"strict-review"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "frontend-tips", "SKILL.md")); err != nil {
		t.Errorf("frontend-tips skill missing")
	}
	if _, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "review-tips", "SKILL.md")); err != nil {
		t.Errorf("review-tips skill missing")
	}
}
```

- [ ] **Schritt 2: Tests ausführen — Fehler erwarten**

```bash
go test ./internal/deploy/...
```

Erwartet: FAIL — `copySkills` existiert noch nicht, Tests schlagen fehl weil Skills nicht kopiert werden.

- [ ] **Schritt 3: `copySkills()`, `copyDir()`, `copyFile()` implementieren**

Füge in `internal/deploy/deploy.go` nach `copyHooks()` hinzu:

```go
func copySkills(sourceDir, destDir string, cfg config.Config) error {
	dirs := []string{
		filepath.Join(sourceDir, "base", "skills"),
		filepath.Join(sourceDir, "profiles", cfg.Profile, "skills"),
	}
	for _, flavor := range cfg.Flavors {
		dirs = append(dirs, filepath.Join(sourceDir, "flavors", flavor, "skills"))
	}

	seen := map[string]bool{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", dir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if seen[name] {
				continue
			}
			seen[name] = true
			src := filepath.Join(dir, name)
			dst := filepath.Join(destDir, ".claude", "skills", name)
			if err := copyDir(src, dst); err != nil {
				return fmt.Errorf("copy skill %s: %w", name, err)
			}
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
```

- [ ] **Schritt 4: `copySkills` in `RunWithClaude()` verdrahten**

In `internal/deploy/deploy.go`, `RunWithClaude()` nach dem `installExtensions`-Block:

```go
	if err := installExtensions(sourceDir, cfg, claudeBin); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	if err := copySkills(sourceDir, destDir, cfg); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
```

- [ ] **Schritt 5: Tests ausführen — alle grün**

```bash
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 6: Commit**

```bash
git add internal/deploy/deploy.go internal/deploy/deploy_test.go
git commit -m "feat(deploy): add copySkills step — distribute skill files to target repos"
```

---

### Task 2: Vier SKILL.md-Dateien anlegen

**Files:**
- Create: `base/skills/release/SKILL.md`
- Create: `base/skills/repo-onboarding/SKILL.md`
- Create: `base/skills/repo-health/SKILL.md`
- Create: `base/skills/forgecrate-advisor/SKILL.md`

Hinweis: Reine Markdown-Inhalte — kein TDD-Zyklus. Korrektheit wird durch die bestehende Test-Suite (kein Regressions-Fehler) und manuelle Sichtprüfung bestätigt.

- [ ] **Schritt 1: `base/skills/release/SKILL.md` anlegen**

```markdown
# Release Skill

Führt durch einen vollständigen Release-Zyklus. Erkennt automatisch das Build-System des Repos.

## Schritte

### 1. Build-System erkennen

Prüfe welche Dateien vorhanden sind:
- `go.mod` → Go: `go test ./...`
- `package.json` → Node: `npm test` oder `yarn test`
- `Cargo.toml` → Rust: `cargo test`
- `Makefile` mit `test`-Target → `make test`
- `pyproject.toml` / `setup.py` → Python: `pytest`

### 2. Tests ausführen

Führe die Test-Suite aus. Bei Fehlern: stoppen, Fehler anzeigen, nicht weiter.

### 3. Changelog prüfen

Prüfe ob `CHANGELOG.md` existiert. Falls ja:
- Ist ein Eintrag für die neue Version vorhanden?
- Falls nein: frage nach der Version und erstelle einen Eintrag unter `## [VERSION] - YYYY-MM-DD`

Falls kein Changelog existiert: hinweisen, aber weiter.

### 4. Version ermitteln

Frage nach der neuen Versionsnummer (SemVer: `vMAJOR.MINOR.PATCH`).

Prüfe ob der Tag schon existiert:
```bash
git tag | grep "^vVERSION$"
```

Falls ja: abbrechen und hinweisen.

### 5. Tag erstellen und pushen

```bash
git tag -a vVERSION -m "Release vVERSION"
git push origin vVERSION
```

### 6. CI prüfen (optional)

Falls `.github/workflows/` existiert:
```bash
gh run list --limit 3
```

Zeige Status der letzten Runs.

## Abschluss

Bestätige Tag und Remote.
```

- [ ] **Schritt 2: `base/skills/repo-onboarding/SKILL.md` anlegen**

```markdown
# Repo Onboarding Skill

Erkundet das Repo und erstellt einen strukturierten Überblick — nützlich beim ersten Einstieg oder nach `forgecrate run`.

## Schritte

### 1. Sprachen und Frameworks erkennen

Prüfe: `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, `pom.xml`, `build.gradle`.

Lese Root-Dateien und die wichtigsten Verzeichnisse.

### 2. Projektstruktur kartieren

- Business-Logik: `internal/`, `src/`, `lib/`, `app/`
- Tests: `*_test.go`, `tests/`, `__tests__/`, `spec/`
- Konfiguration: `.env.example`, `config/`, `*.yaml`
- Externe Abhängigkeiten: aus Lock-Files (`go.sum`, `package-lock.json`, `Cargo.lock`)

### 3. Build & Test verstehen

- `Makefile`: lies die wichtigsten Targets
- `package.json`: lies `scripts`
- `go.mod`: `go build ./...` und `go test ./...`

### 4. Zusammenfassung erstellen und anbieten

Schreibe:

```markdown
## Projektüberblick

**Sprache/Stack:** [...]
**Build:** `[befehl]`
**Test:** `[befehl]`

### Struktur
- `[verzeichnis]/` — [Zweck]

### Wichtige externe Abhängigkeiten
- [Name] — [Zweck]
```

Zeige die Zusammenfassung und frage: "In GENERATED-Block der CLAUDE.md übernehmen?"
```

- [ ] **Schritt 3: `base/skills/repo-health/SKILL.md` anlegen**

```markdown
# Repo Health Skill

Analysiert das Repo und gibt eine priorisierte Liste von Verbesserungsvorschlägen.

## Schritte

### 1. Test-Coverage prüfen

Suche öffentliche Funktionen ohne Tests:
- Go: Exportierte Symbole in `*.go` ohne `*_test.go` Pendant im selben Package
- TypeScript/JS: Exportierte Funktionen ohne `.test.ts` / `.spec.ts`

### 2. Abhängigkeiten prüfen

- Go: `go list -u -m all 2>/dev/null | grep "\["`
- Node: `npm outdated --json 2>/dev/null`
- Rust: `cargo outdated 2>/dev/null`

### 3. Dead Code suchen

- Nicht importierte Packages / nicht aufgerufene Exports
- Längere auskommentierte Code-Blöcke (Hinweis, kein automatisches Löschen)

### 4. Dokumentation prüfen

- Fehlt `README.md`?
- Hat die README einen Getting-Started-Abschnitt?
- Sind öffentliche APIs dokumentiert?

### 5. Sicherheitsmuster prüfen

Suche nach offensichtlichen Problemen:
```bash
grep -rn "password\s*=\s*\"" --include="*.go" --include="*.ts" --include="*.py" .
grep -rn "api_key\s*=\s*\"" --include="*.go" --include="*.ts" --include="*.py" .
```

### 6. Ergebnis ausgeben

Nummerierte Liste nach Priorität:

```
1. [KRITISCH] Hardcoded secret in src/config.go:42
2. [HOCH] 5 exportierte Funktionen ohne Tests in internal/parser/
3. [MITTEL] 3 Dependencies mit Major-Updates verfügbar
4. [NIEDRIG] README fehlt Getting-Started-Abschnitt
```
```

- [ ] **Schritt 4: `base/skills/forgecrate-advisor/SKILL.md` anlegen**

```markdown
# Claude Setup Advisor

Analysiert das aktuelle Repo und empfiehlt das passende forgecrate Profil und Flavor.

## Verfügbare Optionen

**Profile:**
- `backend` — Backend- und CLI-Projekte (Go, Rust, Python, Java, etc.)
- `frontend` — Web-Frontend-Projekte (React, Vue, Angular, Next.js, etc.)
- `fullstack` — Projekte mit Frontend und Backend

**Flavors (kombinierbar):**
- `tdd` — TDD wird strikt eingehalten
- `strict-review` — erhöhte Review-Anforderungen, PR-Pflicht auch bei Solo-Projekten
- `minimal` — kein zusätzlicher Setup, nur Basis-Konfiguration

## Schritte

### 1. Profil bestimmen

Prüfe vorhandene Dateien:
- `package.json` mit `react`, `vue`, `angular`, `next`, `nuxt`, `svelte` in dependencies → `frontend`
- `package.json` ohne Frontend-Framework + Backend-Code → `fullstack`
- `go.mod`, `Cargo.toml`, `pyproject.toml`, `pom.xml` ohne Frontend → `backend`

### 2. Flavors bestimmen

Frage je eine Frage:
1. "Wird TDD in diesem Projekt strikt eingehalten?" → bei Ja: `tdd`
2. "Soll jede Änderung über einen PR gehen?" → bei Ja: `strict-review`

### 3. Empfehlung ausgeben

```
Empfehlung für dieses Repo:

  forgecrate run --profile PROFIL [--flavor FLAVOR]

Begründung:
- Profil PROFIL: [erkannte Signale]
- Flavor FLAVOR: [Grund]
```

Falls forgecrate noch nicht installiert:
```
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest
```
```

- [ ] **Schritt 5: Test-Suite ausführen**

```bash
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 6: Commit**

```bash
git add base/skills/
git commit -m "feat(base): add release, repo-onboarding, repo-health, forgecrate-advisor skills"
```
