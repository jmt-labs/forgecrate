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
- Im Issue einen Kommentar mit kurzem Stand hinterlassen (z.B. "Spec fertig, Plan folgt")

### 3. Plan
- Plan wird in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` geschrieben und committed
- Plan-Pfad wird im gleichen Issue ergänzt (Issue-Body oder Kommentar)
- Im Issue Kommentar: "Plan fertig, Implementierung läuft"

### 4. Implementierung
- Nach jedem abgeschlossenen Task: kurzer Kommentar im Issue
- Bei Abschluss: Issue mit `gh issue close` schließen, letzter Kommentar mit Ergebnis

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
