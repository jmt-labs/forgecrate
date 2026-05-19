#!/usr/bin/env bash
# Erinnerung an Pflicht-Skills — wird bei jeder User-Nachricht ausgegeben.
# Schlank halten: nur wenige Zeilen, vollständig cached nach erster Ausführung.

CFG_FILE=".forgecrate.yaml"
[ -f "$CFG_FILE" ] || CFG_FILE=".claude-setup.yaml"

if [ ! -f "$CFG_FILE" ]; then
  echo "Warnung: Keine State-Datei gefunden (.forgecrate.yaml / .claude-setup.yaml)" >&2
  PROFILE="unbekannt"
  FLAVORS="keine"
elif command -v yq &>/dev/null; then
  PROFILE=$(yq '.profile' "$CFG_FILE" 2>/dev/null) || PROFILE=""
  FLAVORS=$(yq '.flavors | join(", ")' "$CFG_FILE" 2>/dev/null) || FLAVORS=""
  if [ -z "$PROFILE" ] || [ "$PROFILE" = "null" ]; then
    echo "Warnung: .profile fehlt oder leer in $CFG_FILE" >&2
    PROFILE="unbekannt"
  fi
  if [ -z "$FLAVORS" ] || [ "$FLAVORS" = "null" ]; then
    FLAVORS="keine"
  fi
else
  echo "Warnung: yq nicht installiert, YAML-Parsing eingeschränkt" >&2
  PROFILE=$(grep '^profile:' "$CFG_FILE" 2>/dev/null | awk '{print $2}') || PROFILE=""
  FLAVORS=$(grep -A5 'flavors:' "$CFG_FILE" 2>/dev/null | grep '  -' | awk '{print $2}' | tr '\n' ',' | sed 's/,$//') || FLAVORS=""
  if [ -z "$PROFILE" ]; then
    PROFILE="unbekannt"
  fi
  if [ -z "$FLAVORS" ]; then
    FLAVORS="keine"
  fi
fi

echo "## forgecrate — Aktive Konfiguration"
echo "Profil: ${PROFILE} | Flavors: ${FLAVORS}"
echo ""
echo "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs"
