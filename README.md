<div align="center">
  <img src="assets/banner.svg" alt="claude-setup — Reproduzierbares Claude-Setup" width="100%">
</div>

[![Latest Release](https://img.shields.io/github/v/release/jmt-labs/claude-setup)](https://github.com/jmt-labs/claude-setup/releases/latest)
[![CI](https://github.com/jmt-labs/claude-setup/actions/workflows/ci.yml/badge.svg)](https://github.com/jmt-labs/claude-setup/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# claude-setup

**claude-setup** installiert eine reproduzierbare [Claude Code](https://claude.ai/code)-Konfiguration in beliebige Git-Repositories. Ein einzelnes Go-Binary lädt Profile, Flavors, Hooks und MCP-Server-Definitionen von GitHub und komponiert sie über ein Schichtsystem in das Ziel-Repository.

```sh
claude-setup init --profile backend --flavors tdd,strict-review
```

Damit ist Claude Code im Repo sofort einsatzbereit: mit passendem Workflow-Enforcement, Slash-Commands, Branch-Protection und fünf vorintegrierten MCP-Servern — ohne manuelle Konfiguration.

---

## Inhalt

- [Was wird deployt?](#was-wird-deployt)
- [Installation](#installation)
- [Schnellstart](#schnellstart)
- [Das Schichtsystem](#das-schichtsystem)
- [Profile](#profile)
- [Flavors](#flavors)
- [CLI-Referenz](#cli-referenz)
- [Anpassung & lokale Overrides](#anpassung--lokale-overrides)
- [Skills & Slash-Commands](#skills--slash-commands)
- [MCP-Server](#mcp-server)
- [Konflikte beim Update](#konflikte-beim-update)
- [Weiterführende Dokumentation](#weiterführende-dokumentation)
- [Entwicklung & Tests](#entwicklung--tests)

---

## Was wird deployt?

Nach `claude-setup init` liegen folgende Dateien im Repository:

| Datei / Verzeichnis | Zweck |
|---|---|
| `CLAUDE.md` | Verhaltensregeln und Workflow für Claude Code |
| `AGENTS.md` | Richtlinien für Claude Code Agents |
| `.claude/settings.json` | Modell, Hooks, Plugins, Permissions, MCP-Server |
| `.claude/commands/` | Slash-Commands (Skills) |
| `.claude/hooks/` | Pre-Tool- und UserPromptSubmit-Hooks |
| `.claude-setup.yaml` | Deployment-State (Profil, Flavors, Datei-Hashes) |

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew tap jmt-labs/tap
brew install claude-setup
```

### Chocolatey (Windows)

```sh
choco install claude-setup
```

### apt (Ubuntu / Debian)

```sh
curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg

echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/jmt-labs.list

sudo apt update && sudo apt install claude-setup
```

### go install

```sh
go install github.com/jmt-labs/claude-setup/cmd/claude-setup@latest
```

### curl (ohne Paketmanager)

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/claude-setup/main/install.sh | bash
```

Bestimmte Version installieren:

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/claude-setup/main/install.sh | bash -s v1.0.0
```

---

## Schnellstart

**1. Repository initialisieren**

```sh
cd mein-projekt
claude-setup init --profile backend --flavors tdd,strict-review
```

**2. Auf den neuesten Stand aktualisieren**

```sh
claude-setup update
```

**3. Profil wechseln**

```sh
claude-setup update --profile fullstack
```

**4. Verfügbare Optionen einsehen**

```sh
claude-setup list profile    # alle Profile
claude-setup list flavor     # alle Flavors
claude-setup describe profile backend   # Details zu einem Profil
```

---

## Das Schichtsystem

claude-setup komponiert die Konfiguration aus drei übereinanderliegenden Schichten. Jede Schicht erweitert oder überschreibt die darunterliegende:

```
┌─────────────────────────────┐
│  overrides (lokal, optional) │  Höchste Priorität — nie überschrieben
├─────────────────────────────┤
│  flavors   (mehrere wählbar) │  z. B. tdd + strict-review
├─────────────────────────────┤
│  profile   (eines wählbar)   │  z. B. backend
├─────────────────────────────┤
│  base      (immer aktiv)     │  Niedrigste Priorität — immer deployt
└─────────────────────────────┘
```

**Kompositionsregeln:**

| Dateiart | Strategie |
|---|---|
| `CLAUDE.md` / `AGENTS.md` | Textblöcke der Schichten werden aneinandergehängt |
| `.claude/settings.json` | Deep-Merge — tiefere Schichten gewinnen bei Konflikten |
| `.claude/commands/` | Additive Kopie — spätere Schichten überschreiben frühere |
| `.claude/hooks/` | Additive Kopie |

Der Inhalt der Schichten stammt direkt aus diesem Repository unter `base/`, `profiles/` und `flavors/`.

---

## Profile

Ein Profil legt den fachlichen Kontext des Projekts fest. Es ist genau **eines** pro Repository aktiv.

### `backend`

Optimiert für Server-Anwendungen, REST-APIs und Datenbankzugriffe.

- API-Design: REST-First, klar benannte Fehlercodes
- Datenbankzugriffe: typsicher, ausschließlich parametrisierte Queries
- Bevorzugt Integrationstests gegenüber reinen Unit-Tests mit Mocks
- Keine ORM-Magic — explizite Queries sind lesbarer

### `frontend`

Optimiert für komponentenbasierte UI-Entwicklung.

- Komponentenarchitektur und State-Management
- Barrierefreiheit (Accessibility) als First-Class-Anforderung
- Visual Regression und Snapshot Tests

### `fullstack`

Kombiniert Backend- und Frontend-Kontext in einem Profil.

- Shared Types zwischen Client und Server
- End-to-End-Tests über den gesamten Stack
- Klare Grenze zwischen API-Vertrag und Implementierung

---

## Flavors

Flavors ergänzen ein Profil um querschnittliche Praktiken. **Mehrere Flavors** können gleichzeitig aktiv sein.

### `tdd` — Test-Driven Development

Erzwingt den klassischen Red-Green-Refactor-Zyklus:

1. Test schreiben → ausführen (muss scheitern)
2. Minimale Implementierung → Test muss bestehen
3. Refactoring → Tests müssen weiterhin bestehen
4. Committen

Regeln:
- Kein Produktionscode ohne vorherigen Test
- Test-Namen beschreiben Verhalten, nicht Implementierung
- Mocks nur an Systemgrenzen (externe APIs, Datenbanken)
- Vor jedem Bug-Fix: Regressionstest schreiben

### `strict-review` — Pflicht-Code-Review

Erzwingt strukturierte Überprüfung vor jedem Commit:

- Pflichtaufruf von `superpowers:requesting-code-review` vor jedem Commit
- Keine direkten Commits auf `main` / `master`
- PR-Beschreibung enthält: Was geändert wurde, Warum, Wie getestet
- Breaking Changes werden explizit kommuniziert

### `minimal`

Nur grundlegendes Workflow-Enforcement. Geeignet für Projekte, die einen leichten Einstieg bevorzugen.

### `gitops`

Für Infrastructure-as-Code und GitOps-Workflows:

- ArgoCD-App-Topologie wird respektiert
- Clusterweite Regeln (Kyverno, Gatekeeper, RULES.md) werden eingehalten
- Deployments ausschließlich über ArgoCD — keine direkten `kubectl apply`-Kommandos
- Skill: `/claude-setup-gitops-status` für Cluster-Überblick

---

## CLI-Referenz

### `claude-setup init`

Führt eine erstmalige Konfiguration des Repositories durch.

```sh
claude-setup init [--profile <name>] [--flavors <name,name,...>]
```

| Flag | Standard | Beschreibung |
|---|---|---|
| `--profile` | `backend` | Aktives Profil |
| `--flavors` | _(keiner)_ | Kommaseparierte Liste aktiver Flavors |

**Ablauf:** Tarball von GitHub herunterladen → Schichten komponieren → Dateien schreiben → `.claude-setup.yaml` anlegen.

---

### `claude-setup update`

Aktualisiert die Konfiguration auf die neueste Version des Upstream-Repositories.

```sh
claude-setup update [--profile <name>]
```

| Flag | Standard | Beschreibung |
|---|---|---|
| `--profile` | _(aktuell)_ | Profil wechseln |

Bei Konflikten zwischen lokalen Änderungen und dem Upstream wird interaktiv gefragt (siehe [Konflikte beim Update](#konflikte-beim-update)).

---

### `claude-setup list`

Listet alle verfügbaren Profile oder Flavors auf.

```sh
claude-setup list profile
claude-setup list flavor
```

---

### `claude-setup describe`

Zeigt eine detaillierte Beschreibung eines Profils oder Flavors.

```sh
claude-setup describe profile backend
claude-setup describe flavor tdd
```

---

## Anpassung & lokale Overrides

Alle lokalen Anpassungen bleiben bei `claude-setup update` **erhalten**. Es gibt drei Mechanismen:

### CUSTOM-Blöcke in Markdown

In `CLAUDE.md` und `AGENTS.md` können eigene Anweisungen in einen geschützten Block geschrieben werden:

```markdown
<!-- CUSTOM:BEGIN -->
Hier stehen projektspezifische Anweisungen,
die nie vom Update überschrieben werden.
<!-- CUSTOM:END -->
```

### Settings-Overrides

```json
// .claude/overrides/settings.override.json
{
  "permissions": {
    "allow": ["Bash(make:*)"]
  }
}
```

Diese Datei wird beim Merge mit der höchsten Priorität behandelt.

### Eigene Slash-Commands

```
.claude/commands/overrides/mein-command.md
```

Eigene Skills in diesem Verzeichnis ergänzen die verwalteten Commands und werden nie überschrieben.

---

## Skills & Slash-Commands

claude-setup deployt eine Reihe vordefinierter Skills, die direkt in Claude Code als Slash-Commands verfügbar sind:

| Command | Beschreibung |
|---|---|
| `/claude-setup-advisor` | Analysiert das Repo und empfiehlt passendes Profil + Flavors |
| `/claude-setup-repo-onboarding` | Erstellt strukturierten Codebase-Überblick für `CLAUDE.md` |
| `/claude-setup-repo-health` | Findet Verbesserungspotenzial und liefert priorisierte Liste |
| `/claude-setup-test-coverage` | Analysiert Testabdeckung und schlägt nächsten konkreten Test vor |
| `/claude-setup-pr-checklist` | Systematische Überprüfung vor `gh pr create` |
| `/claude-setup-db-migration` | Begleitet Erstellung und Review einer Datenbankmigration |
| `/claude-setup-release` | Führt vollständigen Release-Zyklus durch |

Über die [Superpowers-Skills](https://github.com/anthropics/claude-code-superpowers) stehen zusätzlich Pflicht-Skills zur Verfügung, die automatisch in den Entwicklungs-Workflow eingebunden sind (Brainstorming, TDD, Code-Review, Debugging).

---

## MCP-Server

Die Basis-Konfiguration enthält fünf vorintegrierte MCP-Server:

| Server | Transport | Zweck |
|---|---|---|
| `github` | HTTP (GitHub Copilot) | Issues, PRs, Code-Suche, Branches, Labels |
| `fetch` | stdio (`npx`) | Externe Webinhalte — Dokumentation, RFCs, Changelogs |
| `memory` | stdio (`npx`) | Persistentes projektübergreifendes Wissen (`.claude/memory.json`) |
| `context-mode` | stdio (`npx`) | Automatische Context-Optimierung und Session-History-Suche |
| `context7` | stdio (`npx`) | Aktuelle Bibliotheks-Dokumentation direkt aus den Quell-Repos |

**Voraussetzung für `github`:** Die Umgebungsvariable `GITHUB_PERSONAL_ACCESS_TOKEN` muss gesetzt sein.

**Voraussetzung für stdio-Server:** Node.js / `npx` muss im PATH verfügbar sein.

---

## Konflikte beim Update

Beim `claude-setup update` vergleicht das Tool jede verwaltete Datei mit einem SHA256-Hash, der beim letzten Deploy gespeichert wurde.

Ein **Konflikt** entsteht nur, wenn **beide** Bedingungen zutreffen:
- Die lokale Datei wurde seit dem letzten Deploy manuell verändert **und**
- die neue Upstream-Version weicht von der lokalen Version ab

In diesem Fall erscheint eine interaktive Abfrage:

```
KONFLIKT: .claude/settings.json
  Deine Version: { "model": "claude-opus-4-7", ...
  Neue Version:  { "model": "claude-sonnet-4-6", ...
  [ü]berschreiben / [b]ehalten (Standard: behalten):
```

| Eingabe | Ergebnis |
|---|---|
| `ü` oder `u` | Upstream-Version übernehmen, lokale Änderungen gehen verloren |
| `b` oder Enter | Lokale Version behalten, der lokale Hash wird als neue Basis gespeichert |

**Empfehlung:** Eigene Anpassungen in CUSTOM-Blöcke oder Override-Dateien auslagern — dann entstehen gar keine Konflikte.

---

## Weiterführende Dokumentation

| Thema | Dokument |
|---|---|
| Architektur und Komponenten | [docs/architecture.md](docs/architecture.md) |
| Schichtsystem (Kompositionsregeln) | [docs/layer-system.md](docs/layer-system.md) |
| Abläufe: init, update, enforcement | [docs/flows.md](docs/flows.md) |
| Hook-Referenz | [docs/hooks.md](docs/hooks.md) |
| Profile & Flavors (Details) | [docs/profiles-flavors.md](docs/profiles-flavors.md) |
| Entwicklung & Tests | [docs/development.md](docs/development.md) |

---

## Entwicklung & Tests

**Voraussetzungen:** Go 1.22+, `make`

```sh
# Unit-Tests
make test

# End-to-End-Tests (gegen ein lokales Test-Repository)
make test-e2e

# Codequalität prüfen
make quality

# Binary lokal bauen
make build

# Release (erfordert GoReleaser + GitHub Token)
make release
```

**Repository-Struktur:**

| Pfad | Zweck |
|---|---|
| `cmd/claude-setup/` | CLI-Einstiegspunkt (Cobra) |
| `internal/compose/` | Markdown-, JSON- und Skills-Merge-Logik |
| `internal/deploy/` | Deployment-Orchestrierung und Konflikt-Handling |
| `internal/config/` | `.claude-setup.yaml` lesen/schreiben |
| `internal/github/` | GitHub API Client (Tarball-Download) |
| `internal/extensions/` | Claude Extension Handling |
| `base/` | Basis-Layer — immer deployt |
| `profiles/` | Profil-Layer — je eines wählbar |
| `flavors/` | Flavor-Layer — mehrere kombinierbar |
| `e2e/` | End-to-End-Tests |

Beiträge sind willkommen — bitte einen Issue anlegen und den Entwicklungs-Workflow aus [CLAUDE.md](CLAUDE.md) beachten.

---

## Lizenz

[MIT](LICENSE) — Copyright (c) jmt-labs
