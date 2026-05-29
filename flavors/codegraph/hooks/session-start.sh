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

( set -C; : > "$STAMP.lock" ) 2>/dev/null || exit 0

if [ ! -d "$REPO_ROOT/.codegraph" ]; then
  ( cd "$REPO_ROOT" && codegraph init -i && touch "$STAMP"; rm -f "$STAMP.lock" ) >/dev/null 2>&1 &
else
  ( cd "$REPO_ROOT" && codegraph index && touch "$STAMP"; rm -f "$STAMP.lock" ) >/dev/null 2>&1 &
fi
