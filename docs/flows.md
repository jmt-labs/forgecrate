# Abläufe

## `forgecrate init`

```
forgecrate init --profile backend --flavors tdd
        │
        ├── .forgecrate.yaml lesen (falls bereits vorhanden → Migration)
        ├── GitHub-Tarball nach tmpDir herunterladen (internal/github)
        ├── Layer komponieren: base → profile → flavors (internal/compose)
        │       ├── CLAUDE.md / AGENTS.md   → Markdown-Konkat + CUSTOM-Block-Schutz
        │       ├── .claude/settings.json   → Deep-JSON-Merge
        │       ├── .claude/commands/       → additive Skill-Kopie
        │       └── .claude/hooks/          → additive Hook-Kopie
        ├── Extensions installieren (internal/extensions)
        │       ├── Plugins via `claude plugin install --scope project <source>`
        │       │       oder `claude plugin marketplace add <source>` (method: marketplace)
        │       └── `.mcp.json` aus extensions.yaml generieren (MCP-Server-Konfiguration)
        ├── memory-bank/ scaffolden (nur fehlende Dateien aus base/memory-bank/ kopieren)
        ├── SHA256-Hashes der deployten Dateien speichern
        ├── .forgecrate.yaml schreiben (profile, flavors, permission_mode, hashes)
        └── Done.
```

Der alte Alias `forgecrate run` ist als Backwards-Compat noch aktiv.

## `forgecrate update`

```
forgecrate update [--profile <neu>]
        │
        ├── .forgecrate.yaml lesen (Fehler wenn nicht vorhanden)
        ├── Profil ggf. überschreiben (--profile)
        ├── GitHub-Tarball nach tmpDir herunterladen
        ├── Layer rekompositionieren (overrides/ bleibt unangetastet)
        ├── Pro Datei:
        │       ├── SHA256 der lokalen Datei berechnen
        │       ├── vergleichen mit gespeichertem Hash aus .forgecrate.yaml
        │       ├── + vergleichen mit Upstream-Hash
        │       └── bei echtem Konflikt: interaktiv [o]verwrite / [k]eep
        ├── Neue Hashes speichern
        └── Done.
```

## `forgecrate config`

```
forgecrate config
        │
        ├── .forgecrate.yaml lesen (Fehler wenn nicht vorhanden)
        ├── Profile + Flavors aus GitHub-Source listen
        ├── TUI (Charmbracelet Huh): Profil-Select + Flavor-MultiSelect
        │       (aktuelle Auswahl vorausgewählt)
        ├── Deploy mit neuer Auswahl (Update-Flow)
        └── Done.
```

## `forgecrate set-permission-mode`

```
forgecrate set-permission-mode <bypass|plan|ask|auto>
        │
        ├── Modus validieren (internal/config.ValidatePermissionMode)
        ├── .forgecrate.yaml lesen
        ├── .claude/settings.json patchen (deploy.PatchPermissionMode)
        └── .forgecrate.yaml mit neuem permission_mode schreiben
```

## Enforcement zur Laufzeit

```
User schreibt Prompt
        │
        ├── UserPromptSubmit-Hook: prompt-submit.sh
        │       └── ruft `forgecrate hook prompt-submit` → gibt aktives Profil +
        │           Pflicht-Skill-Liste aus
        │
        ├── Claude liest Prompt + CLAUDE.md (Pflicht-Skills-Tabelle)
        ├── Claude ruft relevanten Skill auf (brainstorming, tdd, etc.)
        │
        ├── PreToolUse-Hook (vor Bash/Edit/Write): pre-tool.sh
        │       ├── auf `main` werden destruktive Bash-Kommandos blockiert
        │       └── kontextabhängige Erinnerung an relevante Pflicht-Skills
        │
        └── Tool wird ausgeführt
```

Details zu den Hooks: [docs/hooks.md](hooks.md).
