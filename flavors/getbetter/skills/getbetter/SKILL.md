# GETBETTER

Reflektiert die aktuelle Session und aktualisiert `.claude/GETBETTER.md` mit synthetisierten Erkenntnissen.

## Ablauf

1. **Bestehende Erkenntnisse laden**

   Prüfe ob `.claude/GETBETTER.md` existiert:
   ```bash
   cat .claude/GETBETTER.md 2>/dev/null || echo "(keine bisherigen Erkenntnisse)"
   ```

2. **Aktuelle Session reflektieren**

   Analysiere die aktuelle Session und extrahiere Erkenntnisse in diesen Kategorien:

   - **Entscheidungen** — Was wurde entschieden und warum? Welche Alternativen wurden verworfen?
   - **Anti-Patterns** — Was lief schief? Was hätte früher erkannt werden sollen?
   - **Was funktioniert** — Welche Ansätze haben sich bewährt? Was sollte beibehalten werden?

   Formuliere frei — kein starres Format, aber bleib konkret und präzise.

3. **Synthetisieren und schreiben**

   Führe bestehende und neue Erkenntnisse zusammen:
   - Bestehende Punkte bleiben erhalten, sofern sie nicht durch neue überholt werden
   - Überschneidende Punkte werden verdichtet, nicht doppelt geführt
   - Neue Erkenntnisse werden eingearbeitet

   Schreibe das Ergebnis nach `.claude/GETBETTER.md`:

   ```markdown
   # GETBETTER

   _Letzte Aktualisierung: YYYY-MM-DD_

   ## Entscheidungen
   [synthetisierter Inhalt]

   ## Anti-Patterns
   [synthetisierter Inhalt]

   ## Was funktioniert
   [synthetisierter Inhalt]
   ```

4. **Bestätigen**

   Gib eine kurze Zusammenfassung: wie viele Punkte wurden hinzugefügt, geändert, verdichtet.
