# Hooks

forgecrate deployt zwei Hook-Scripts nach `.claude/hooks/` im Ziel-Repo. Die
Scripts liegen ursprünglich in `base/hooks/` und werden bei jedem
`forgecrate update` überschrieben.

## UserPromptSubmit — `prompt-submit.sh`

Wird bei jeder User-Nachricht ausgeführt, bevor Claude den Prompt verarbeitet.

**Verhalten:** ruft `forgecrate hook prompt-submit` auf. Der Helper liest
`.forgecrate.yaml` und gibt aus:

```
## forgecrate — Aktive Konfiguration
Profil: backend | Flavors: tdd, strict-review

Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs
Recherche-Pflicht (erzwungen): einmal pro Session vor dem ersten Edit/Write WebSearch/context7/fetch nutzen — nicht raten (Block via pre-tool Hook).
```

Die letzte Zeile (Recherche-Pflicht) entfällt automatisch, wenn der Flavor
`no-research` aktiv ist.

## PreToolUse — `pre-tool.sh`

Wird vor `Bash`, `Edit`, `Write` und `MultiEdit` ausgeführt. Der Hook liest das
PreToolUse-JSON über **stdin** (inkl. `tool_name`, `tool_input`, `transcript_path`).

**Verhalten:**

- **Warnt** auf Branch `main`/`master` bei destruktiven Bash-Kommandos:
  `git commit`, `git push`, `git reset --hard`, sowie Schreib-Redirektionen
  (`>`, `>>`) auf versionierte Dateien.
- **Warnt bei fehlender Recherche** via `forgecrate hook require-research`:
  `Edit`/`Write`/`MultiEdit` erzeugen eine Warnung, solange im gesamten Session-Transcript
  kein assistant-`tool_use` mit einem Recherche-Tool
  (`WebSearch`, `WebFetch`, `mcp__fetch__*`, `mcp__context7__*`) steht.
  Eine Recherche pro Session schaltet weitere Warnungen für die Session ab.
  - Flavor `no-research` deaktiviert die Warnung vollständig.
  - Flavor `force-research` erweitert sie auf schreibende Bash-Befehle
    (`sed -i`, `tee`, `dd of=`, Redirects außerhalb `/tmp`).
  - **Fail-open**: fehlt die Binary, das Transcript oder ist es unparsebar, wird
    nicht gewarnt (verhindert Dauerstörungen).
- **Erinnert** kontextabhängig an relevante Pflicht-Skills — z. B. an
  `superpowers:test-driven-development`, wenn Code-Dateien editiert werden.

Wichtig: der Hook ist eine **lokale** Schutzschicht. GitHub Branch Protection
Rules müssen serverseitig zusätzlich konfiguriert sein, um direkte Pushes auch
dort zu unterbinden.

## Hooks anpassen oder erweitern

Die deployten Dateien unter `.claude/hooks/` werden bei `forgecrate update`
überschrieben. Lokale Anpassungen gehen verloren.

**Empfohlener Workflow:**

1. Eigene Hooks unter anderem Dateinamen ablegen (z. B.
   `.claude/hooks/team-custom.sh`)
2. In `.claude/settings.json` (CUSTOM-Block der Settings) zusätzliche
   `hooks`-Einträge ergänzen, die diese eigenen Scripts referenzieren

So bleiben die forgecrate-Hooks aktuell, eigene Logik überlebt Updates.

## Quellen

- Hook-Scripts: `base/hooks/prompt-submit.sh`, `base/hooks/pre-tool.sh`
- Helper-Binary: `cmd/forgecrate/hook.go` (`forgecrate hook prompt-submit`,
  `forgecrate hook require-research`)
