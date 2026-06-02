---
name: forgecrate-roadmap-triage
description: >-
  PFLICHT nach jedem abgeschlossenen Brainstorming. Bewertet das Ergebnis mit
  WSJF-Score und K.O.-Kriterien. Entscheidet: "Umsetzbar jetzt" (→ writing-plans)
  oder "Future Feature" (→ GitHub Issue anlegen). UNBEDINGT nutzen, sobald eine
  neue Idee, ein Bug, eine Erweiterung oder ein Feature-Wunsch erwähnt wird, oder
  Fragen wie "ist das wichtig?", "gehört das auf die Roadmap?", "Bugfix oder
  Feature?", "was als nächstes?", "Release planen", "Backlog aufräumen" gestellt
  werden — auch wenn der Skill nicht ausdrücklich genannt wird.
---

# roadmap-triage

**System of Record: GitHub Issues.** Kein Markdown, keine lokalen Dateien als Datenspeicher. Alles lebt in Issues.

## Voraussetzungen

```bash
gh auth status     # muss ok sein — sonst STOP, User bitten: ! gh auth login
gh repo view       # muss Ziel-Repo erkennen
```

Fehlt `gh` oder fehlt Auth → **sofort stoppen**. Niemals selbst einloggen, keine Tokens anfordern oder entgegennehmen.

## Operating-Prinzipien

1. **GitHub Issues = einzige Quelle der Wahrheit.** Capture = sofort ein Issue.
2. **Aktiver Meilenstein ist heilig.** Neue Ideen gehen per Default ins Backlog.
3. **Blocker-Test als Gate.** In aktiven Milestone nur, wenn DoD ohne diese Idee nicht erreichbar ist.
4. **Nichts wird gelöscht.** Nur schließen: `shipped` (completed) oder `dropped` (not planned + Grund).
5. **Schnell entscheiden.** Eine WSJF-Zahl schlägt jede Diskussion.
6. **WIP-Limit = 7** offene Items pro Milestone (Default). Voll → raus bevor rein.

## Labels einmalig sicherstellen

```bash
bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh setup-labels
```

Legt alle `stage:*`, `type:*`, `prio:*` und `dropped` idempotent an. **Einmal vor erstem Einsatz ausführen.**

## Modi

| Modus | Trigger | Schnellbefehl |
|---|---|---|
| **Capture** | Idee/Bug/Feature erwähnt | `bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh capture "<titel>"` |
| **Triage** | „Backlog aufräumen", Inbox verarbeiten | `bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh inbox` |
| **Plan Release** | „Release planen", neuer Meilenstein | `bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh backlog-ranked` |
| **Groom** | „Backlog reviewen", WSJF veraltet | `bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh resurface` |
| **Status** | „Was liegt an?", „Wie steht's?" | `bash .claude/skills/forgecrate-roadmap-triage/roadmap.sh status` |

Detaillierter Ablauf je Modus → [references/modes.md](references/modes.md)

## Issue-Body-Template

```
**WSJF:** value=_ · time-crit=_ · risk-opp=_ · size=_ → **score=_**
**Resurface:** — | <tag/Bedingung/Datum>

<Beschreibung / Kontext, 1–3 Sätze>

**Definition of Done:** <sobald eingeplant, aus dem Meilenstein abgeleitet>
```

WSJF-Scoring → [references/wsjf-scoring.md](references/wsjf-scoring.md)  
Datenmodell (Labels, Milestones, Stages) → [references/data-model.md](references/data-model.md)

## Sicherheitsregeln

Vor diesen Aktionen **immer Bestätigung einholen**:
- Massen-Edits (> 1 Issue gleichzeitig)
- Schließen eines Issues
- Verschieben in den aktiven Meilenstein
- Anlegen mehrerer Issues auf einmal

## Entscheidung nach Brainstorming

**PFLICHT:** Direkt nach jedem abgeschlossenen `superpowers:brainstorming` diesen
Entscheidungsbaum durchlaufen. Kein Plan, keine Implementierung ohne dieses Gate.

### K.O.-Kriterien (überwiegen WSJF-Score)

Falls eines zutrifft → sofort **Future Feature**, unabhängig vom Score:

1. **Fehlende Abhängigkeiten** — externes Service, anderes Feature oder Infrastruktur noch nicht fertig
2. **Scope zu groß** — nicht in einem PR + Review-Zyklus abschließbar (>3 Tage Arbeit)
3. **Kein definierbarer Nutzer-Impact** — Akzeptanzkriterien nicht formulierbar

### WSJF-Schwelle

- Score ≥ 2.0 **und** kein K.O.-Kriterium → **Umsetzbar jetzt**
- Score < 2.0 **oder** K.O.-Kriterium → **Future Feature**

### Ausgabe

**Umsetzbar jetzt:**
```
✅ Umsetzbar jetzt (WSJF: X.X)
→ Weiter mit writing-plans
```

**Future Feature:**
```
📋 Future Feature (WSJF: X.X | K.O.: <Grund oder "keines">)
→ GitHub Issue anlegen mit Label "future-feature"
→ Kein Plan, keine Implementierung jetzt
```

### Trigger

Wird aufgerufen:
- **Pflicht** — direkt nach jedem abgeschlossenen `superpowers:brainstorming`
- **Automatisch** — wenn eine neue Idee, ein Bug, eine Erweiterung oder Feature-Wunsch erwähnt wird
- **Auf Anfrage** — bei Fragen wie "ist das wichtig?", "gehört das auf die Roadmap?", "machen wir das?"
