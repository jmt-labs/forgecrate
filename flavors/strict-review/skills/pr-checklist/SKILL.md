# PR-Checkliste

Systematische Überprüfung vor `gh pr create`. Alle offenen Punkte müssen adressiert sein.

## Ablauf

1. **Breaking Changes identifizieren**
   - Geänderte Funktionssignaturen: `git diff main -- '*.go' '*.ts' '*.py'` nach `func ` / `export` suchen
   - Entfernte oder umbenannte Felder in structs/types
   - Datenbankmigrationen vorhanden?

2. **Testabdeckung für geänderte Dateien**
   ```bash
   git diff main --name-only | grep -v '_test\.' | while read f; do
     base="${f%.*}"
     test_candidates=("${base}_test.go" "${base}.test.ts" "tests/test_${base##*/}.py")
     for t in "${test_candidates[@]}"; do [ -f "$t" ] && echo "OK: $t" && break; done
   done
   ```
   Dateien ohne zugehörigen Test: explizit auflisten.

3. **Dokumentation aktuell?**
   - `README.md` — deckt neue Features/Flags ab?
   - `CLAUDE.md` — GENERATED-Block aktuell?
   - Inline-Kommentare für nicht-offensichtliches Verhalten?

4. **PR-Beschreibung vollständig?** Prüfe ob folgende Punkte beantwortet sind:
   - Was wurde geändert?
   - Warum (Ticket/Issue verlinkt)?
   - Wie wurde getestet?

5. **Checkliste ausgeben**

   ```
   ✅ Keine Breaking Changes erkannt
   ⚠️  internal/deploy/deploy.go hat keinen zugehörigen Test für copyFile
   ✅ README.md aktuell
   ❌ PR-Beschreibung: "Wie getestet" fehlt
   ```

   Offene ❌-Punkte müssen behoben sein bevor `gh pr create` aufgerufen wird.
