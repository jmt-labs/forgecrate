# Accessibility Audit

Prüft Barrierefreiheit in geänderten UI-Komponenten.

## Ablauf

1. **Geänderte Dateien ermitteln**
   ```bash
   git diff --name-only | grep -E '\.(tsx?|jsx?|vue|svelte|html)$'
   ```
   Falls keine UI-Dateien geändert wurden: "Keine UI-Dateien geändert — Audit übersprungen."

2. **Für jede Datei prüfen:**

   **Fehlende `alt`-Attribute:**
   ```bash
   grep -n '<img' <datei> | grep -v 'alt='
   ```

   **Interaktive Elemente ohne Label:**
   ```bash
   grep -n '<button\|<a ' <datei> | grep -v 'aria-label\|aria-labelledby'
   ```
   Prüfe zusätzlich ob der Button/Link sichtbaren Textinhalt hat (kein reines Icon ohne Label).

   **Formular-Inputs ohne `<label>`:**
   ```bash
   grep -n '<input\|<textarea\|<select' <datei> | grep -v 'type="hidden"'
   ```
   Für jeden Fund: prüfe ob ein `<label for="...">` mit passendem `id` existiert.

   **Inline-Farbkontrast-Warnung:**
   ```bash
   grep -n 'color:.*#[fF][fF]\|color:.*white\|color:.*rgb(255' <datei>
   ```
   Hinweis wenn helle Textfarbe auf weißem Hintergrund vermutet wird.

3. **Ausgabe** — Befunde mit Datei und Zeile:
   ```
   components/Button.tsx:12  <img> ohne alt-Attribut
   components/Nav.tsx:34     <button> ohne aria-label und ohne sichtbaren Text
   ```
   Bei keinen Befunden: "Keine Barrierefreiheitsprobleme in geänderten Dateien gefunden."

## Hinweise

- Kein Ersatz für echte Screenreader-Tests — deckt die häufigsten statischen Fehler ab.
- Bei generierten Dateien (`.min.js`, Build-Output) überspringen.
