#!/usr/bin/env bash
# Erinnerung an Pflicht-Skills — wird bei jeder User-Nachricht ausgegeben.
# Schlank halten: nur wenige Zeilen, vollständig cached nach erster Ausführung.

CFG_FILE=".forgecrate.yaml"
[ -f "$CFG_FILE" ] || CFG_FILE=".claude-setup.yaml"

PROFILE=$(grep 'profile:' "$CFG_FILE" 2>/dev/null | awk '{print $2}')
FLAVORS=$(grep -A5 'flavors:' "$CFG_FILE" 2>/dev/null | grep '  -' | awk '{print $2}' | tr '\n' ',' | sed 's/,$//')

echo "## forgecrate — Aktive Konfiguration"
echo "Profil: ${PROFILE:-unbekannt} | Flavors: ${FLAVORS:-keine}"
echo ""
echo "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs"
