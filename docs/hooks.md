# Hooks

## UserPromptSubmit — `prompt-submit.sh`

Wird bei jeder User-Nachricht ausgeführt.

**Output:** Aktives Profil + Pflicht-Skill-Liste (wenige Zeilen, gecacht).

## PreToolUse — `pre-tool.sh`

Wird vor `Bash`, `Edit`, `Write` ausgeführt.

**Input:** `$CLAUDE_TOOL_NAME` (Tool-Name)

**Output:** Kontextabhängige Erinnerung an relevante Pflicht-Skills.

## Override

Hooks können in `overrides/settings.override.json` ergänzt oder ersetzt werden.
