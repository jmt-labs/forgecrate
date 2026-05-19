# Abläufe

## init-Flow

```
forgecrate init --profile backend --flavors tdd
        │
        ├── .forgecrate.yaml lesen (falls vorhanden)
        ├── GitHub tarball downloaden → tmpDir
        ├── Layer compositionieren: base → profile → flavors
        │       ├── CLAUDE.md: MergeMarkdown(layers, existing)
        │       ├── AGENTS.md: MergeMarkdown(layers, existing)
        │       ├── settings.json: DeepMergeJSON(base, profile, overrides)
        │       └── commands/: MergeSkills(srcDirs, dest)
        ├── Hooks nach .claude/hooks/ kopieren
        ├── .forgecrate.yaml schreiben
        └── Done.
```

## update-Flow

```
forgecrate update [--profile <p>]
        │
        ├── .forgecrate.yaml lesen (Fehler wenn nicht vorhanden)
        ├── Profile überschreiben wenn --profile angegeben
        ├── GitHub tarball downloaden → tmpDir
        ├── Layer rekompositionieren (overrides/ unangetastet)
        └── Done.
```

## Enforcement-Flow

```
User schreibt Prompt
        │
        ├── UserPromptSubmit-Hook: prompt-submit.sh
        │       └── Gibt Profil + Pflicht-Skills aus
        │
        ├── Claude liest Prompt + CLAUDE.md-Pflicht-Skills-Tabelle
        ├── Claude ruft relevanten Skill auf
        │
        ├── PreToolUse-Hook (vor Bash/Edit/Write): pre-tool.sh
        │       └── Gibt kontextabhängige Erinnerung aus
        │
        └── Tool wird ausgeführt
```
