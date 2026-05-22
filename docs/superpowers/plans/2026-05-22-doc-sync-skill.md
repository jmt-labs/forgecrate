# forgecrate-doc-sync Skill — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Einen neuen forgecrate-Skill `forgecrate-doc-sync` erstellen, der alle Docs im Repo via Git-Delta + KI-Analyse mit dem aktuellen Code-Stand abgleicht und veraltete Abschnitte direkt aktualisiert.

**Architecture:** Reiner Markdown-Skill (keine Go-Code-Änderungen). Der Skill liegt in `base/skills/forgecrate-doc-sync/SKILL.md` und wird beim nächsten `forgecrate update` in Ziel-Repos deployed. Er ist in drei Phasen gegliedert: Git-Delta-Filter → KI-Analyse → Gezieltes Rewrite.

**Tech Stack:** Markdown (SKILL.md), Bash-Snippets im Skill-Body, Claude Code Edit/Read-Tools für den Rewrite-Schritt.

**Spec:** `docs/superpowers/specs/2026-05-22-doc-sync-design.md`

---

## File Structure

| Datei | Aktion | Verantwortung |
|---|---|---|
| `base/skills/forgecrate-doc-sync/SKILL.md` | Erstellen | Vollständiger Skill-Body |
| `flavors/strict-review/skills/forgecrate-pr-checklist/SKILL.md` | Bereits geändert | Schritt 3 ruft `/forgecrate-doc-sync` auf |

---

### Task 1: SKILL.md erstellen

**Files:**
- Create: `base/skills/forgecrate-doc-sync/SKILL.md`

- [ ] **Schritt 1: Verzeichnis anlegen**

```bash
mkdir -p base/skills/forgecrate-doc-sync
```

Erwartung: kein Fehler, Verzeichnis existiert.

- [ ] **Schritt 2: SKILL.md mit vollständigem Inhalt schreiben**

Datei `base/skills/forgecrate-doc-sync/SKILL.md` mit folgendem Inhalt anlegen:

```markdown
# Doc Sync

Gleicht alle Dokumentationsdateien im Repo mit dem aktuellen Code-Stand ab
und aktualisiert veraltete Abschnitte direkt. Kein interaktiver Modus —
der Skill arbeitet durch und liefert am Ende einen kompakten Report.

## Scope

| Doku-Typ | Verhalten |
|---|---|
| `docs/*.md` | Analysiert; veraltete Abschnitte werden neu geschrieben |
| `README.md` | Wie `docs/*.md`; Fokus auf Installation, Konfiguration, Beispiele |
| `CHANGELOG.md` | Kein Auto-Edit; Warnung wenn Commits seit letztem Eintrag existieren |
| `*.go` GoDoc | Exportierte Symbole ohne Kommentar werden ergänzt; veraltete Kommentare korrigiert |

## Ablauf

### Phase 1 — Git-Delta-Filter

Für jede Doku-Datei den Zeitstempel des letzten Commits ermitteln und
Code-Dateien identifizieren die danach geändert wurden:

```bash
for doc in docs/*.md README.md; do
  [ -f "$doc" ] || continue
  doc_ts=$(git log -1 --format="%ct" -- "$doc" 2>/dev/null || echo 0)
  changed=$(git log --format="%H" -- '*.go' | while read hash; do
    commit_ts=$(git show -s --format="%ct" "$hash")
    [ "$commit_ts" -gt "$doc_ts" ] && git diff-tree --no-commit-id -r --name-only "$hash" -- '*.go'
  done | sort -u)
  [ -n "$changed" ] && echo "VERDÄCHTIG: $doc" && echo "$changed"
done
```

GoDoc-Filter — exportierte Symbole in `.go`-Dateien die neuer sind als
ihre Doku-Kommentare:

```bash
find . -name "*.go" -not -path "./.git/*" | while read f; do
  file_ts=$(git log -1 --format="%ct" -- "$f" 2>/dev/null || echo 0)
  grep -n "^func [A-Z]\|^type [A-Z]\|^var [A-Z]\|^const [A-Z]" "$f" | while IFS=: read lineno content; do
    prev_line=$(( lineno - 1 ))
    comment=$(sed -n "${prev_line}p" "$f")
    echo "$comment" | grep -q "^//" || echo "KEIN GODOC: $f:$lineno — $content"
  done
done
```

Dateien die älter als 90 Tage sind aber keine verdächtigen Code-Änderungen
haben, separat markieren:

```bash
NOW=$(date +%s)
for doc in docs/*.md README.md; do
  [ -f "$doc" ] || continue
  doc_ts=$(git log -1 --format="%ct" -- "$doc" 2>/dev/null || echo 0)
  age=$(( (NOW - doc_ts) / 86400 ))
  [ "$age" -gt 90 ] && echo "ALT (${age}d): $doc"
done
```

### Phase 2 — KI-Analyse

Für jedes verdächtige Code/Doku-Paar (aus Phase 1):

1. Code-Datei lesen (Read-Tool)
2. Korrespondierende Doku-Datei lesen (Read-Tool)
3. Inhaltliche Abweichungen identifizieren:
   - Veraltete Beschreibungen (Funktion umbenannt, Signatur geändert)
   - Fehlende Abschnitte für neue Features oder Flags
   - Nicht mehr existierende Symbole oder Kommandos in Beispielen

**Mapping Code → Doku** (forgecrate-spezifisch):

| Code-Pfad | Betroffene Doku |
|---|---|
| `internal/compose/` | `docs/layer-system.md` |
| `internal/deploy/` | `docs/flows.md`, `docs/architecture.md` |
| `internal/extensions/` | `docs/architecture.md` |
| `cmd/forgecrate/` | `README.md`, `docs/flows.md` |
| `profiles/`, `flavors/` | `docs/profiles-flavors.md` |
| `base/hooks/` | `docs/hooks.md` |

Für Repos ohne dieses Mapping: inhaltliche Zuordnung über Schlagwörter
(Paket-Name, Typ-Bezeichnung, Kommando-Name) in Doku und Code vergleichen.

### Phase 3 — Gezieltes Rewrite

Für jeden als veraltet identifizierten Abschnitt:

- Nur den betroffenen Abschnitt neu schreiben (Edit-Tool)
- Niemals die ganze Datei ersetzen — `old_string` muss den exakten
  veralteten Abschnitt enthalten, `new_string` den aktualisierten
- Bestehende Überschriften, Formatierung und Dateistruktur beibehalten
- Sprache: automatisch aus bestehendem Dateiinhalt erkennen und beibehalten
- Ton: präzise und technisch, keine Marketing-Sprache
- Länge: so kurz wie möglich ohne Informationsverlust

GoDoc-Kommentare direkt in `.go`-Dateien ergänzen oder korrigieren:

```
// FunctionName beschreibt in einem Satz was die Funktion tut.
func FunctionName(...) ...
```

CHANGELOG.md nicht editieren — stattdessen Commit-Anzahl seit letztem Tag ausgeben:

```bash
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
  COUNT=$(git log "${LAST_TAG}..HEAD" --oneline | wc -l | tr -d ' ')
  echo "CHANGELOG.md: $COUNT Commits seit $LAST_TAG"
else
  echo "CHANGELOG.md: kein Tag gefunden — manuell prüfen"
fi
```

## Ausgabe

```
## Doc Sync — Ergebnis

### Aktualisiert
- docs/architecture.md — Abschnitt "Architektur" (internal/deploy geändert)
- README.md — Abschnitt "Installation" (cmd/forgecrate geändert)
- internal/deploy/deploy.go — GoDoc für CopyDir, MergeFiles

### Keine Änderung nötig
- docs/layer-system.md
- docs/profiles-flavors.md

### Hinweise
- CHANGELOG.md: 12 Commits seit v0.9.0
- docs/hooks.md: keine verdächtigen Code-Änderungen, aber 95 Tage alt — manuell prüfen empfohlen
```
```

- [ ] **Schritt 3: Datei prüfen**

```bash
cat base/skills/forgecrate-doc-sync/SKILL.md | head -5
```

Erwartung: erste Zeile ist `# Doc Sync`.

---

### Task 2: Smoke-Test — Skill auf dem forgecrate-Repo ausführen

**Files:**
- Read: `base/skills/forgecrate-doc-sync/SKILL.md` (der Skill selbst)

- [ ] **Schritt 1: Skill im aktuellen Repo aufrufen**

`/forgecrate-doc-sync` im laufenden Claude-Code-Terminal eingeben und ausführen.

- [ ] **Schritt 2: Output-Format verifizieren**

Erwarteter Output enthält mindestens:
- Abschnitt `### Aktualisiert` oder `### Keine Änderung nötig`
- Abschnitt `### Hinweise` mit CHANGELOG-Warnung (da Commits seit letztem Tag vorhanden)
- Kein Crash, keine leere Ausgabe

- [ ] **Schritt 3: Bei Abweichungen SKILL.md nachbessern**

Konkrete Muster-Probleme und Fixes:
- Bash-Snippet funktioniert nicht → Kommando debuggen und im Skill korrigieren (Edit-Tool)
- Output-Format weicht ab → Ausgabe-Abschnitt im Skill anpassen
- Phase 2 Mapping fehlt für ein Repo → Hinweis in Skill ergänzen: "Für Repos ohne dieses Mapping..."

---

### Task 3: Committen

**Files:**
- Commit: `base/skills/forgecrate-doc-sync/SKILL.md`

- [ ] **Schritt 1: Datei stagen**

```bash
git add base/skills/forgecrate-doc-sync/SKILL.md
```

- [ ] **Schritt 2: Status prüfen**

```bash
git status
```

Erwartung: `base/skills/forgecrate-doc-sync/SKILL.md` unter "Changes to be committed".

- [ ] **Schritt 3: Committen**

```bash
git commit -m "feat(skills): forgecrate-doc-sync Skill hinzufügen

Gleicht docs/*.md, README.md, CHANGELOG.md und GoDoc mit aktuellem
Code-Stand ab. Drei Phasen: Git-Delta-Filter, KI-Analyse, Rewrite.
Pflichtschritt in forgecrate-pr-checklist (strict-review Flavor).

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

Erwartung: Commit-Hash wird ausgegeben, kein Fehler.
