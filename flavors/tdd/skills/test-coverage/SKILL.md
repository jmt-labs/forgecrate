# Test Coverage

Analysiert Testabdeckung und schlägt den nächsten konkreten Test vor.

## Ablauf

1. **Coverage-Report erzeugen** — erkenne Build-System:
   - Go: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`
   - Node/Jest: `npm test -- --coverage 2>/dev/null || npx jest --coverage`
   - Python: `pytest --cov --cov-report=term-missing 2>/dev/null`

2. **Lücken auflisten** — filtere Zeilen mit `0.0%` Coverage und zeige die Top 5:
   ```
   internal/deploy/deploy.go:copyFile     0.0%
   internal/config/config.go:Validate     0.0%
   ...
   ```

3. **Nächsten Test vorschlagen** — für die oberste Lücke (höchste Zeilenzahl ohne Coverage):
   - Funktion/Methode benennen
   - Sinnvollen Testfall formulieren: Eingabe, Erwartung, Randfall
   - Dateiname und Testfunktionsname vorschlagen

   Beispiel:
   ```
   Vorschlag: internal/deploy/deploy_test.go
   func TestCopyFileCreatesParentDirs(t *testing.T) {
     // prüft ob copyFile fehlende Parent-Verzeichnisse anlegt
   }
   ```

4. **Fragen** — "Soll ich diesen Test anlegen?"

## Hinweise

- Läuft nach jedem Feature-Zyklus, nicht nur bei roten Tests.
- Fokus auf Verhalten, nicht auf Line-Coverage als Selbstzweck.
