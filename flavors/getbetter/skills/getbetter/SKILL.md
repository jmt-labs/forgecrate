# GETBETTER

Reflektiert die aktuelle Session und speichert synthetisierte Erkenntnisse im memory MCP.

## Ablauf

1. **Bestehende Erkenntnisse laden**

   `mcp__memory__read_graph` aufrufen. Entities vom Typ `session-reflection` auslesen.
   Falls keine vorhanden: starte mit einer leeren Basis.

   Falls keine verwertbare Session-History vorhanden ist (z.B. direkt nach Sessionbeginn oder nach Kompaktierung): kurze Meldung ausgeben und Skill beenden — keine Erkenntnisse erzwingen.

2. **Aktuelle Session reflektieren**

   Analysiere die aktuelle Session und extrahiere Erkenntnisse in diesen Kategorien:

   - **Entscheidungen** — Was wurde entschieden und warum? Welche Alternativen wurden verworfen?
   - **Anti-Patterns** — Was lief schief? Was hätte früher erkannt werden sollen?
   - **Was funktioniert** — Welche Ansätze haben sich bewährt? Was sollte beibehalten werden?

   Formuliere frei — kein starres Format, aber bleib konkret und präzise.

3. **Synthetisieren und speichern**

   Führe bestehende und neue Erkenntnisse zusammen:
   - Bestehende Punkte bleiben erhalten; ein Punkt gilt als überholt wenn ein neuer denselben Sachverhalt präziser beschreibt oder ihn explizit widerlegt
   - Überschneidende Punkte werden verdichtet, nicht doppelt geführt
   - Neue Erkenntnisse werden eingearbeitet

   `mcp__memory__add_observations` aufrufen mit:
   - `entityName`: `session-reflection`
   - `contents`: Liste der synthetisierten Erkenntnisse (je Erkenntnis ein String)

   Format pro Erkenntnis:
   ```
   [YYYY-MM-DD] <Kategorie>: <Erkenntnis in einem Satz>
   ```

   Kategorien: `workflow`, `tooling`, `pattern`, `mistake`, `decision`.

   Kein GETBETTER.md schreiben — memory MCP ist das einzige Speicherziel.

4. **Bestätigen**

   Gib eine kurze Zusammenfassung: wie viele Punkte wurden hinzugefügt, verdichtet oder entfernt.
