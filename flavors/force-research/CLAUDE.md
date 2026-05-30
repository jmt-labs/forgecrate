## Force-Research-Flavor (verschärfte Recherche-Pflicht)

Dieser Flavor verschärft die ohnehin erzwungene Recherche-Pflicht des base layer:

- Der harte PreToolUse-Block (kein Edit/Write/MultiEdit ohne vorherige Recherche
  einmal pro Session) gilt **zusätzlich für schreibende Bash-Befehle** — auch
  Datei-Schreibzugriffe via Shell (`sed -i`, `tee`, `dd of=`, Redirects außerhalb
  `/tmp`) werden ohne vorherige Recherche blockiert. Damit ist die Umgehung „Datei per
  Shell schreiben statt Edit/Write" geschlossen.
- Kein impliziter Ausnahmefall. Bewusster Verzicht ausschließlich über den Flavor
  `no-research`.

Die Durchsetzung liegt vollständig im base-Hook (`pre-tool.sh` →
`forgecrate hook require-research`); dieser Flavor aktiviert lediglich die
zusätzliche Bash-Prüfung über die aktive Konfiguration.
