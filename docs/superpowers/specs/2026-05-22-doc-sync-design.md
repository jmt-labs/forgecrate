# Doc Sync — Design

**Datum:** 2026-05-22  
**Status:** Abgestimmt  
**Skill-Name:** `forgecrate-doc-sync`

## Ziel

Ein forgecrate-Skill der bei Bedarf alle Dokumentationsdateien im Repo mit dem aktuellen Code-Stand abgleicht und veraltete Abschnitte direkt aktualisiert. Der Skill arbeitet ohne interaktive Rückfragen durch und liefert am Ende einen kompakten Report.

## Scope

| Doku-Typ | Verhalten |
|---|---|
| `docs/*.md` | Vollständig analysiert; veraltete Abschnitte werden neu geschrieben |
| `README.md` | Wie `docs/*.md`; Fokus auf Installations-, Konfigurations- und Beispiel-Abschnitte |
| `CHANGELOG.md` | Kein Auto-Edit; Warnung wenn Commits seit letztem Eintrag existieren |
| `*.go` GoDoc | Exportierte Symbole ohne Doc-Kommentar werden ergänzt; veraltete Kommentare werden korrigiert |

## Ablauf

### Phase 1 — Git-Delta-Filter

Für jede Doku-Datei wird der Zeitstempel des letzten Commits ermittelt (`git log -1 --format="%ct" -- <file>`). Code-Dateien die danach geändert wurden, gelten als "verdächtig". Ergebnis: Mapping `code-file → betroffene Doku-Datei(en)`.

Für GoDoc: exportierte Symbole in `.go`-Dateien deren letzter Code-Commit neuer ist als der letzte Kommentar-Stand.

### Phase 2 — KI-Analyse

Für jedes verdächtige Code/Doku-Paar:
- Aktuellen Code-Abschnitt lesen
- Korrespondierenden Doku-Abschnitt lesen
- Inhaltliche Abweichungen identifizieren: veraltete Beschreibungen, fehlende neue Features, falsche Beispiele, nicht mehr existierende Symbole

Dateien ohne verdächtige Code-Änderungen, die aber älter als 90 Tage sind, werden als "manuell prüfen empfohlen" gemeldet — nicht automatisch editiert.

### Phase 3 — Gezieltes Rewrite

Nur veraltete Abschnitte werden neu geschrieben — nie die ganze Datei. Bestehende Überschriften, Formatierung und Dateistruktur bleiben erhalten.

GoDoc-Kommentare werden direkt in den `.go`-Dateien aktualisiert (via Edit-Tool).

`CHANGELOG.md` wird nicht editiert.

## Schreibstil

- **Sprache**: Automatisch erkannt aus bestehendem Datei-Inhalt — Skill schreibt konsequent in der bereits verwendeten Sprache weiter
- **Ton**: Präzise und technisch; keine Marketing-Sprache
- **Länge**: So kurz wie möglich ohne Informationsverlust

## Output-Format

```
## Doc Sync — Ergebnis

### Aktualisiert
- docs/architecture.md — Abschnitt "Architektur" (internal/deploy neu)
- README.md — Abschnitt "Installation" (cmd/forgecrate geändert)
- internal/deploy/deploy.go — GoDoc für CopyDir, MergeFiles

### Keine Änderung nötig
- docs/layer-system.md
- docs/profiles-flavors.md

### Hinweise
- CHANGELOG.md: 12 Commits seit letztem Eintrag (letzter: v0.9.0)
- docs/hooks.md: keine verdächtigen Code-Änderungen, aber >90 Tage alt — manuell prüfen empfohlen
```

## Nicht-Ziele

- Kein automatisches Befüllen von `CHANGELOG.md`
- Kein Rewrite von Spec- oder Plan-Dokumenten unter `docs/superpowers/`
- Keine strukturellen Änderungen an Doku-Dateien (Abschnitte umbenennen, restrukturieren)
- Kein Linting oder Stilprüfung jenseits des inhaltlichen Abgleichs
