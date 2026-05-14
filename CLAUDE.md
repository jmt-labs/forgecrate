# Claude-Konfiguration — claude-setup Repo

## Entwicklungs-Workflow

Dieser Workflow gilt für alle Features, Bugfixes und sonstigen Änderungen:

### 1. Brainstorming
- Vor jeder Implementierung: `superpowers:brainstorming` aufrufen
- Ergebnis: abgestimmtes Design

### 2. Spec
- Vor dem ersten Commit: Branch anlegen (`git checkout -b feat/<thema>`)
- Alle weiteren Arbeiten (Spec, Plan, Implementierung) laufen auf diesem Branch
- Spec wird nach Brainstorming in `docs/superpowers/specs/YYYY-MM-DD-<thema>-design.md` geschrieben und committed
- Spec wird einem GitHub-Issue zugeordnet:
  - Existierendes Issue suchen (`gh issue list`) → verlinken
  - Sonst: neues Issue anlegen (`gh issue create`) mit Spec-Pfad im Body
- Branch-Name im Issue vermerken: "Branch: `feat/<thema>`"
- Im Issue einen Kommentar mit kurzem Stand hinterlassen (z.B. "Spec fertig, Plan folgt")

### 3. Plan
- Plan wird in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` geschrieben und committed
- Plan-Pfad wird im gleichen Issue ergänzt (Issue-Body oder Kommentar)
- Im Issue Kommentar: "Plan fertig, Implementierung läuft"

### 4. Implementierung
- Nach jedem abgeschlossenen Task: kurzer Kommentar im Issue

### 5. PR & Abschluss
- PR erstellen (`gh pr create`), Issue im PR-Body verlinken (z.B. "Closes #1")
- Issue wird **erst nach dem Merge des PR** geschlossen — nicht früher
- GitHub schließt das Issue automatisch beim Merge wenn "Closes #N" im PR-Body steht

### Ticket-Kommentare
Immer kurz halten — ein Satz reicht:
- "Spec geschrieben: `docs/superpowers/specs/...`"
- "Plan fertig: `docs/superpowers/plans/...`"
- "Task 2/4 abgeschlossen"
- "Fertig, merged in main"

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer über Branch + PR, nie direkt auf `main`
