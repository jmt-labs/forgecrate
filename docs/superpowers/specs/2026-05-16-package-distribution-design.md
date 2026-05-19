# Spec: Package Distribution via GoReleaser

**Datum:** 2026-05-16
**Thema:** Homebrew, Chocolatey und Ubuntu apt-Repository via GoReleaser

## Ziel

Den Release-Workflow um drei Paketformate erweitern, sodass Nutzer `forgecrate` über ihren nativen Paketmanager installieren können:

- **Homebrew** (`brew install jmt-labs/tap/forgecrate`) — macOS & Linux
- **Chocolatey** (`choco install forgecrate`) — Windows, Community-Repo
- **apt** (`apt install forgecrate`) — Ubuntu/Debian via GitHub-Pages-hosted apt-Repository

## Architektur

### Neue Repositories

| Repository | Zweck |
|---|---|
| `jmt-labs/homebrew-tap` | Homebrew-Formula; GoReleaser pusht bei jedem Release |
| `jmt-labs/apt` | GitHub Pages apt-Repository; separater CI-Job aktualisiert es |

### CI-Secrets

| Secret | Verwendung |
|---|---|
| `HOMEBREW_TAP_TOKEN` | GitHub PAT mit Write-Zugriff auf `homebrew-tap` und `apt` |
| `CHOCOLATEY_API_KEY` | Chocolatey Community API Key |
| `GPG_PRIVATE_KEY` | GPG-Privatschlüssel für apt-Repo-Signierung |
| `GPG_KEY_ID` | Key-ID des GPG-Schlüssels |

## Komponenten

### 1. GoReleaser-Konfiguration (`.goreleaser.yaml`)

Ersetzt `make release`. Verantwortlich für:

- **Cross-Compilation** — identische Targets wie bisher (`linux_amd64`, `linux_arm64`, `windows_amd64`, `windows_arm64`, `darwin_arm64`)
- **GitHub Release** — lädt alle Binaries und das `.deb` als Assets hoch
- **Homebrew-Formula** — pusht aktualisierte Formula in `jmt-labs/homebrew-tap` via `HOMEBREW_TAP_TOKEN`
- **Chocolatey** — baut `.nupkg` und publiziert es auf `push.chocolatey.org` via `CHOCOLATEY_API_KEY`
- **`.deb`-Paket** — wird via `nfpm` gebaut (`/usr/local/bin`, Maintainer: Markus Hartmann)

### 2. Makefile

```makefile
release:
    goreleaser release --clean
```

Das bisherige manuelle `GOOS=... go build`-Target entfällt. GoReleaser legt sein eigenes `dist/` an.

### 3. Release-Workflow (`.github/workflows/release.yml`)

**Job `release`** (bestehend, angepasst):
- `make test`, `make test-e2e` bleiben
- `make release` → `goreleaser release --clean`
- Benötigt Secrets: `HOMEBREW_TAP_TOKEN`, `CHOCOLATEY_API_KEY`, `GPG_PRIVATE_KEY`, `GPG_KEY_ID`

**Job `publish-apt`** (neu, `needs: release`):
1. Checkout `jmt-labs/apt` (via `HOMEBREW_TAP_TOKEN`)
2. `.deb`-Assets vom GitHub Release herunterladen (`gh release download`)
3. `reprepro includedeb stable *.deb` — fügt Pakete in das Repository ein
4. GPG-Key importieren, `reprepro export` — signiert `Release`-Datei
5. Commit + Push → GitHub Pages veröffentlicht automatisch

### 4. apt-Repository-Struktur (`jmt-labs/apt`, Branch `gh-pages`)

```
conf/
  distributions          # reprepro-Konfiguration (initial committen)
pool/main/c/forgecrate/
  forgecrate_<version>_amd64.deb
  forgecrate_<version>_arm64.deb
dists/stable/
  Release, Release.gpg, InRelease
  main/binary-amd64/Packages, Packages.gz
  main/binary-arm64/Packages, Packages.gz
KEY.gpg                  # Public Key für Nutzer
```

## Einmalige Voraussetzungen (manuell, vor erstem Release)

1. **`jmt-labs/homebrew-tap`** anlegen (leer, public)
2. **`jmt-labs/apt`** anlegen; GitHub Pages auf Branch `gh-pages` aktivieren; `conf/distributions` und `KEY.gpg` initial committen
3. **GPG-Schlüsselpaar** generieren (`rsa4096`); Public Key als `KEY.gpg` committen; Private Key als Secret `GPG_PRIVATE_KEY` hinterlegen; Key-ID als `GPG_KEY_ID`
4. **Chocolatey-Account** anlegen, API-Key als Secret `CHOCOLATEY_API_KEY` hinterlegen
5. **GitHub PAT** mit Write-Rechten auf beide Repos als `HOMEBREW_TAP_TOKEN` hinterlegen

## Nutzer-Installationsanleitung (nach Release)

```bash
# Homebrew (macOS/Linux)
brew tap jmt-labs/tap
brew install forgecrate

# Chocolatey (Windows)
choco install forgecrate

# Ubuntu/Debian
curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/jmt-labs.list
sudo apt update && sudo apt install forgecrate
```

## User-facing Dokumentation (README.md)

Der bestehende README-Installations-Abschnitt wird grundlegend überarbeitet: klare Tabs/Blöcke pro Paketmanager, copy-paste-ready, englische Sprache (da internationales Publikum).

**Neue Struktur des `## Installation`-Abschnitts:**

```markdown
## Installation

### Homebrew (macOS / Linux)
brew tap jmt-labs/tap
brew install forgecrate

### Chocolatey (Windows)
choco install forgecrate

### apt (Ubuntu / Debian)
curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] \
  https://jmt-labs.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/jmt-labs.list
sudo apt update && sudo apt install forgecrate

### go install
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest

### curl (ohne Paketmanager)
curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash
```

**Weitere README-Verbesserungen:**
- Kurze Erklärung direkt unter dem Titel (was ist es, warum braucht man es — 2–3 Sätze, Englisch)
- Quick-Start-Block bleibt, wird aber nach dem Installations-Abschnitt platziert
- Versions-Badge (GoReleaser generiert automatisch einen `latest`-Tag, den wir als Shield einbinden)

## Nicht im Scope

- macOS `darwin_amd64` (Intel) — war bisher nicht gebaut, bleibt außen vor
- Launchpad PPA — bewusst durch GitHub-Pages-apt-Repo ersetzt (einfacher, gleiche UX)
- RPM / AUR / Snap — kein Bedarf geäußert
