#!/usr/bin/env bash
# Codegraph-Index bei Session-Start im Hintergrund aktualisieren.
# Läuft via prompt-submit.sh Auto-Discovery bei UserPromptSubmit.

if ! command -v codegraph >/dev/null 2>&1; then
  exit 0
fi

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
if [ -z "$REPO_ROOT" ]; then
  exit 0
fi

codegraph index "$REPO_ROOT" &>/dev/null &
