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

HEAD="$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null)"
STAMP="/tmp/codegraph-indexed-${HEAD}"
if [ -f "$STAMP" ]; then
  exit 0
fi
touch "$STAMP"

codegraph index "$REPO_ROOT" &>/dev/null &
