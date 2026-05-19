# Hooks

## UserPromptSubmit — `prompt-submit.sh`

Wird bei jeder User-Nachricht ausgeführt.

**Output:** Aktives Profil + Pflicht-Skill-Liste (wenige Zeilen, gecacht).

## PreToolUse — `pre-tool.sh`

Wird vor `Bash`, `Edit`, `Write` ausgeführt.

**Input:** `$CLAUDE_TOOL_NAME` (Tool-Name)

**Output:** Kontextabhängige Erinnerung an relevante Pflicht-Skills.

## Hooks anpassen

Hooks liegen nach dem Deployment unter `.claude/hooks/` im Ziel-Repo und können dort direkt bearbeitet werden. Sie werden bei `forgecrate update` überschrieben — lokale Anpassungen in `.claude/hooks/` vorher sichern.
