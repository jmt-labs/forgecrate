# Base Layer: Team-Rollen & Multi-Agent-Konfiguration — Design

**Datum:** 2026-05-14

## Ziel

Die `base/CLAUDE.md` und `base/.claude/settings.json` erhalten eine klare Multi-Agent-Konfiguration, die Entwickler-Teamrollen auf Superpowers-Skills und Modelle abbildet. Der Hauptagent wird auf `claude-sonnet-4-6` gesetzt.

## Designentscheidungen

**Hauptagent-Modell in `settings.json`:** Das Feld `model` legt das Standardmodell für die Hauptsession fest. Subagenten können davon abweichen.

**Team-Rollen-Tabelle in `CLAUDE.md`:** Eine Tabelle im GENERATED-Block bildet sieben Rollen eines kleinen Entwicklerteams auf Superpowers-Skills, Modelle und Effort-Level ab. Der Hauptagent koordiniert und kann bei Bedarf eigenständig davon abweichen.

**DevOps und Security:** Bleiben separate Flavors — nicht im base layer.

## Rollen-Mapping

| Rolle | Superpowers-Skill | Modell | Effort |
|---|---|---|---|
| Analyst / Product Owner | `superpowers:brainstorming` | `claude-opus-4-7` | high |
| Tech Lead / Architekt | `superpowers:writing-plans` | `claude-opus-4-7` | high |
| Entwickler | `superpowers:test-driven-development` | `claude-sonnet-4-6` | medium |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `claude-haiku-4-5-20251001` | low |
| Reviewer | `superpowers:requesting-code-review` | `claude-sonnet-4-6` | medium |
| QA / Abschluss | `superpowers:verification-before-completion` | `claude-sonnet-4-6` | medium |
| Debugger | `superpowers:systematic-debugging` | `claude-sonnet-4-6` | medium |

## Änderungen

### `base/CLAUDE.md`

Neuer Abschnitt `## Team-Rollen & Subagent-Konfiguration` wird im GENERATED-Block nach dem bestehenden Inhalt eingefügt.

### `base/.claude/settings.json`

Feld `"model": "claude-sonnet-4-6"` wird als erstes Feld hinzugefügt.
