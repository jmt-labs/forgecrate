#!/usr/bin/env bash
# Wird vor Bash/Edit/Write aufgerufen.
# Gibt eine Warnung aus wenn keine Session-Notiz über einen aufgerufenen Skill existiert.
# Claude sieht diese Ausgabe als Kontext.

TOOL="${CLAUDE_TOOL_NAME:-}"

case "$TOOL" in
  Edit|Write)
    echo "## Pre-Tool Check"
    echo "Du verwendest $TOOL. Stelle sicher:"
    echo "- superpowers:brainstorming wurde für neue Features aufgerufen"
    echo "- superpowers:test-driven-development wurde vor der Implementierung aufgerufen"
    ;;
  Bash)
    echo "## Pre-Tool Check"
    echo "Du verwendest Bash. Stelle sicher dass destruktive Aktionen mit dem User abgestimmt sind."
    ;;
esac
