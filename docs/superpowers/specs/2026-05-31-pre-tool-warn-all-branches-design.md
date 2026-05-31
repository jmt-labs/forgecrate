# Design: pre-tool.sh — Warnungen auf allen Branches

**Datum:** 2026-05-31

## Ziel

Destruktive Bash-Befehle sollen auf **allen** Branches erkannt werden — nicht nur auf `main`/`master`. Auf Feature-Branches gibt der Hook eine Warnung aus (Claude sieht sie, läuft aber weiter). Auf `main`/`master` bleibt das harte Blocken erhalten.

## Verhalten nach Änderung

| Situation | Ergebnis |
|---|---|
| `Edit`/`Write`/`MultiEdit` auf `main`/`master` | `continue: false` — geblockt |
| `Edit`/`Write`/`MultiEdit` auf anderem Branch | OK-Hinweis (unverändert) |
| Destruktiver Bash-Befehl auf `main`/`master` | `continue: false` — geblockt |
| Destruktiver Bash-Befehl auf anderem Branch | `hookSpecificOutput` mit Warnung — läuft weiter |
| Normaler Bash-Befehl auf beliebigem Branch | OK-Hinweis (unverändert) |

## Destruktive Muster (unverändert)

- `git commit` direkt
- `git push --force` / `git push -f`
- `git push origin main` / `git push origin master`
- `git reset --hard`
- `git clean -f`
- Schreib-Redirectionen (`>`, `>>`) in nicht-`/tmp/`-Pfade

## Umsetzung

`base/hooks/pre-tool.sh`, Bash-Zweig: Die destruktiven Pattern-Checks werden aus dem `if [ "$BRANCH" = "main" ]`-Block herausgezogen. Vor jedem Check wird branching eingebaut:
- `main`/`master` → `continue: false`
- anderer Branch → `hookSpecificOutput` mit Warnung

Der `.forgecrate.yaml`-Hash für `pre-tool.sh` wird nach dem Update synchronisiert.

## Nicht im Scope

- Konfigurierbarkeit der Warnschwelle
- Neue destruktive Muster
