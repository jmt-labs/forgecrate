#!/usr/bin/env bash
# Erinnerung an Pflicht-Skills — wird bei jeder User-Nachricht ausgegeben.
# Schlank halten: nur wenige Zeilen, vollständig cached nach erster Ausführung.

PROFILE=$(grep 'profile:' .claude-setup.yaml 2>/dev/null | awk '{print $2}')
FLAVORS=$(grep -A5 'flavors:' .claude-setup.yaml 2>/dev/null | grep '  -' | awk '{print $2}' | tr '\n' ',' | sed 's/,$//')

echo "## Claude Setup — Aktive Konfiguration"
echo "Profil: ${PROFILE:-unbekannt} | Flavors: ${FLAVORS:-keine}"
echo ""
echo "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs"
