#!/usr/bin/env bash
# PreToolUse-Hook. stdin (tool_name, tool_input, transcript_path) einmal puffern
# und an beide forgecrate-Unterkommandos weiterreichen.

STDIN_JSON=""
if [ ! -t 0 ]; then
  STDIN_JSON=$(cat)
fi

if command -v forgecrate >/dev/null 2>&1; then
  # Destruktive-Befehl-Prüfung: blockt auf main, warnt auf anderen Branches
  OUT=$(printf '%s' "$STDIN_JSON" | forgecrate hook pre-tool)
  if [ -n "$OUT" ]; then
    printf '%s' "$OUT"
    # Bei hartem Block (continue:false) sofort beenden
    if printf '%s' "$OUT" | grep -q '"continue":false'; then
      exit 0
    fi
  fi

  # Recherche-Pflicht: blockt Edit/Write/MultiEdit ohne vorherige Recherche
  DECISION=$(printf '%s' "$STDIN_JSON" | forgecrate hook require-research)
  if [ -n "$DECISION" ]; then
    printf '%s' "$DECISION"
  fi
fi
# Hinweis: Dieser Hook ist keine alleinige Schutzschicht. GitHub Branch Protection ergänzen.
