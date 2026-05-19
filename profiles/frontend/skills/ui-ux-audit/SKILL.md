# UI/UX Audit

Tiefgehender Review der gesamten Web-UI eines Repos aus Sicht eines UI/UX-Designers. Identifiziert Bugs, UX-Friction, visuelle Inkonsistenzen, A11y-Lücken und Performance-Probleme. Liefert kleinteilige, ausformulierte GitHub-Issues, die andere Agenten parallel abarbeiten können.

Ergänzt `accessibility-audit` (schnelle statische Checks pro Datei) durch einen ganzheitlichen Audit über alle UI-Bereiche, gegliedert nach User-Journeys.

## Wann verwenden

- Vor einem Major-Release als UI-Reife-Check
- Bei wachsender UI-Komplexität (>2k Zeilen, mehrere Views)
- Auf Wunsch des Nutzers ("UI prüfen", "UI-Audit", "Review der Web-UI")
- Nach größeren Frontend-Refactorings

Nicht für Single-File-Änderungen — dafür ist `accessibility-audit` oder `pr-checklist` zuständig.

## Ablauf

### 1. Scope abstimmen

Frage den Nutzer (mit `AskUserQuestion`, falls verfügbar) drei kurze Klärungen, sofern nicht offensichtlich:

1. **Analyse-Tiefe**: Code-only oder Code + lokaler Build mit Browser-Verifikation?
2. **Issue-Granularität**: Sehr kleinteilig (1 Fix = 1 Issue) oder gebündelt (1 Thema = 1 Issue)?
3. **Schwerpunkt-Bereiche**: Auth, Hauptworkflow, Konfiguration, Querschnitt (Mehrfachauswahl).

Bei kleinen UIs (<1k Zeilen) sind die Defaults: Code-only, gebündelt, alle Bereiche.

### 2. UI-Struktur erfassen

Identifiziere die UI-Architektur:

```bash
# Frontend-Hauptverzeichnis finden
find . -type d \( -name 'web' -o -name 'frontend' -o -name 'ui' -o -name 'app' \) -not -path '*/node_modules/*' | head -5

# Größe und Komponentenstruktur
find <ui-root> -type f \( -name '*.tsx' -o -name '*.ts' -o -name '*.jsx' -o -name '*.js' -o -name '*.vue' -o -name '*.svelte' -o -name '*.html' -o -name '*.css' \) -not -path '*/node_modules/*' -not -path '*/dist/*' | xargs wc -l | sort -rn | head -20
```

Notiere: Framework (oder Vanilla), Build-Setup, Datei-Hotspots, CSS-Architektur (Tokens? Inline? Tailwind?), Routing.

### 3. Parallele Bereichs-Audits

Wenn das Agent-Tool verfügbar ist: starte bis zu 4 parallele Audit-Agents, je einer pro Bereich. Jeder bekommt einen klaren Scope:

- **Auth/Onboarding**: Login, Signup, Logout, Setup-Flows
- **Hauptworkflow**: Kern-Feature der App (z.B. Sessions, Dashboard, Editor)
- **Wizards/Dialoge**: Multi-Step-Forms, Modals, Confirm-Dialogs, Command-Palette
- **Konfiguration/Settings**: Drawer, Settings-Pages, Admin, Profile

Jeder Agent prüft **fünf Kategorien**:

| Kategorie | Beispiele |
|---|---|
| **Bug** | Race-Conditions, Memory-Leaks, Form-Reset stale, Double-Submit, falsches Re-Rendering, tote Buttons |
| **UX** | Discoverability, fehlende Tastatur-Shortcuts, fehlende Empty-/Loading-/Error-States, unklare Validierung, fehlende CTAs, "Type-to-Confirm" für irreversible Aktionen |
| **Visual** | Spacing-Inkonsistenzen, Border/Radius-Verstöße gegen Token, Truncation bei langen Werten, Button-Reihenfolge, Hover-/Focus-Affordances |
| **A11y** | `role`, `aria-*`, Focus-Management (Modal-Open/Close), Tabs-Pattern, Splitter-Tastatur, Form-Labels, `outline:none`-Verstöße, autocomplete, Caps-Lock |
| **Performance** | innerHTML-Re-Renders ganzer Listen, fehlende Event-Delegation, Polling+SSE-Doppelungen, fehlendes Debouncing, fehlende EventSource-Cleanups |

### 4. Findings konsolidieren

Aus den Bereichs-Audits ergeben sich typischerweise 50–80 Roh-Findings. Konsolidiere:

- **Dubletten zusammenfassen** (Double-Submit kommt oft in mehreren Bereichen vor → 1 Querschnitt-Issue)
- **Trivia bündeln** (3 LOW-Polish-Items zu einem Issue, wenn aus demselben Bereich)
- **Severity vergeben**: HIGH (Bug/Datenverlust/A11y-Blocker), MEDIUM (Friction), LOW (Polish)

Ziel: 20–35 kleinteilige Issues. Bei kleinen UIs entsprechend weniger.

### 5. Issues anlegen

**Master-Issue zuerst** mit Checkliste aller Sub-Issues:

```markdown
# [UI/UX Audit] <Datum> — Master

## Findings nach Schweregrad
| Schweregrad | Anzahl |
|---|---|
| HIGH | X |
| MEDIUM | Y |
| LOW | Z |

## Sub-Issues nach Bereich
### Auth (n)
- [ ] #N · A1 · [HIGH][Bug] <Titel>
...

## Empfohlene Reihenfolge
1. HIGH-Bugs zuerst
2. HIGH-A11y
3. MEDIUM-Performance
4. MEDIUM-UX
5. LOW-Polish
```

**Sub-Issues** mit einheitlichem Body:

```markdown
## Problem
**Schweregrad:** <high|medium|low>
**Kategorie:** <Bug | UX | A11y | Performance | Visual>

<Beobachtung in 1–3 Sätzen>

**Ort:** `<datei>:<zeilen>`

## Risiko
<Wer ist betroffen? Was kann konkret passieren?>

## Fix
<Konkreter Vorschlag, möglichst mit Code-Snippet>

## Akzeptanzkriterien
- [ ] <prüfbar>
- [ ] <prüfbar>

## Referenz
Part of UI/UX Audit <Datum> (Master-Issue: #<MASTER>)
```

**Labels**: `ui` (immer), zusätzlich `ux` / `a11y` / `bug` / `performance` je nach Kategorie, plus `severity:high|medium|low`. Fehlende Labels vor dem ersten Issue-Anlegen (falls möglich) im Repo erstellen.

**Titel-Konvention**: `[UI/UX][<Bereich>] <prägnanter Titel>` — macht Filtern leicht.

### 6. Master-Issue final updaten

Nach dem Anlegen aller Sub-Issues das Master-Issue mit den vergebenen Nummern verlinken (Checkliste mit `#N`-Referenzen).

## Verifikation

Vor dem Issue-Anlegen — falls möglich — die UI bauen, um Build-Fehler auszuschließen:

```bash
# Beispiele
npm --prefix <ui-root> run build
pnpm --filter <ui-root> build
```

Bei Build-Fehler den Audit pausieren und den Fehler dem Nutzer melden.

## Hinweise

- **Keine Code-Änderungen** während des Audits — das ist die Aufgabe der nachgelagerten Issue-Bearbeitung.
- **Branch-Disziplin**: Audit-Issues haben keinen Branch; Bearbeitung pro Issue erfolgt separat auf eigenem Branch.
- **TUI- oder Design-System-Constraints** aus `CLAUDE.md` / `AGENTS.md` respektieren — kein generischer Marketing-/Dashboard-Vorschlag, wenn der Style explizit minimalistisch ist.
- **Backend-Themen** (CSRF, Auth-Cookies, Crypto) explizit ausschließen — die gehören in `code-review` / Security-Audits, nicht hier.
- **Bei sehr großen UIs** (>10k Zeilen): pro Bereich einen eigenen Master-Issue, statt ein einzelner Mega-Master.
