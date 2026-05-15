#!/usr/bin/env bash
TOOL="${CLAUDE_TOOL_NAME:-}"

case "$TOOL" in
  Edit|Write|MultiEdit)
    BRANCH=$(git branch --show-current 2>/dev/null)
    if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
      printf '{"continue": false, "stopReason": "Direkte Änderungen auf main sind verboten. Branch anlegen: git checkout -b feat/<thema> — dann erst Edit/Write verwenden."}'
      exit 0
    fi
    echo '{"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": "Branch-Check OK. Stelle sicher: brainstorming und tdd Skills wurden aufgerufen."}}'
    ;;
  Bash)
    echo '{"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": "Bash-Aufruf. Keine destruktiven Aktionen ohne Bestätigung."}}'
    ;;
esac
