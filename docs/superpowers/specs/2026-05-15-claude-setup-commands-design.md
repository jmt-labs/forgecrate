# Design: Slash-Commands für claude-setup-Skills

## Ziel

Alle Skills, die `claude-setup run` installiert, sollen als Slash-Commands im Format
`/claude-setup-<name>` aufrufbar sein. Das ermöglicht schnellen Zugriff ohne den
genauen Skill-Namen kennen zu müssen.

## Hintergrund

Claude Code lädt `.md`-Dateien aus `~/.claude/commands/` als Slash-Commands.
`claude-setup` hat in `compose.go` bereits `composeSkills`, das Dateien aus
`base/.claude/commands/`, `profiles/<p>/.claude/commands/` und
`flavors/<f>/.claude/commands/` nach `<destDir>/.claude/commands/` merged.

Das Verzeichnis `base/.claude/commands/` existiert noch nicht und ist leer.

## Design

### Kein Umbau — nur neue Dateien

Skills bleiben in ihren bestehenden Verzeichnissen (`base/skills/release/`,
`flavors/tdd/skills/test-coverage/`, etc.). Es werden nur Command-Wrapper-Dateien
hinzugefügt.

### Command-Dateien in `base/.claude/commands/`

Alle Commands kommen in `base/.claude/commands/`, unabhängig von Profil oder Flavor.
Wenn ein Skill für das aktuelle Profil/Flavor nicht installiert wurde, meldet der
Skill-Tool einen klaren Fehler — das ist akzeptables Verhalten.

| Command-Datei | Slash-Command | Ruft auf |
|---|---|---|
| `claude-setup-advisor.md` | `/claude-setup-advisor` | `Skill("claude-setup-advisor")` |
| `claude-setup-release.md` | `/claude-setup-release` | `Skill("release")` |
| `claude-setup-repo-health.md` | `/claude-setup-repo-health` | `Skill("repo-health")` |
| `claude-setup-repo-onboarding.md` | `/claude-setup-repo-onboarding` | `Skill("repo-onboarding")` |
| `claude-setup-db-migration.md` | `/claude-setup-db-migration` | `Skill("db-migration")` |
| `claude-setup-test-coverage.md` | `/claude-setup-test-coverage` | `Skill("test-coverage")` |
| `claude-setup-pr-checklist.md` | `/claude-setup-pr-checklist` | `Skill("pr-checklist")` |
| `claude-setup-accessibility-audit.md` | `/claude-setup-accessibility-audit` | `Skill("accessibility-audit")` |
| `claude-setup-github-release.md` | `/claude-setup-github-release` | `Skill("github-release")` |

### Dateiformat

Jede Command-Datei ist minimal:

```markdown
---
description: <Kurzbeschreibung aus dem SKILL.md>
---

Use the Skill tool to invoke the "<skill-name>" skill.
```

### Deploy-Pfad

`claude-setup run` → `compose.Run()` → `composeSkills()` → merged
`base/.claude/commands/*.md` nach `~/.claude/commands/` → Slash-Commands aktiv.

Kein Umbau an `deploy.go`, `compose.go` oder `extensions.yaml` nötig.

### Zukünftiger Plugin-Schritt (nicht Teil dieses Features)

Später wird das Plugin-System genutzt, um den Namespace `claude-setup:<name>`
(mit Doppelpunkt) zu ermöglichen. Die Command-Dateien können dann 1:1 übernommen werden.

## Nicht in scope

- Umbenennung bestehender Skill-Verzeichnisse
- Änderungen an `deploy.go` oder `compose.go`
- Profile- oder Flavor-spezifische Commands
- Plugin-Registrierung
