# Package Distribution via GoReleaser — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the release pipeline with Homebrew, Chocolatey (community repo) and Ubuntu apt (GitHub Pages) package distribution via GoReleaser.

**Architecture:** GoReleaser replaces the manual `make release` cross-compilation target and handles Homebrew formula updates and Chocolatey publishing natively. A second CI job (`publish-apt`) downloads the generated `.deb` from the GitHub release, uses `reprepro` to insert it into a GPG-signed apt repository hosted on GitHub Pages in the `jmt-labs/apt` repo.

**Tech Stack:** Go · GoReleaser v2 · nfpm (`.deb`) · reprepro · GitHub Actions · GitHub Pages

---

## File Structure

| File | Action | Purpose |
|---|---|---|
| `.goreleaser.yaml` | Create | GoReleaser config: builds, nfpm, Homebrew, Chocolatey |
| `Makefile` | Modify | Replace manual release target with `goreleaser release --clean` |
| `LICENSE` | Create | Required by Chocolatey publisher |
| `.github/workflows/release.yml` | Modify | Use GoReleaser, add `publish-apt` job |
| `README.md` | Modify | English rewrite, installation sections for all package managers |
| `jmt-labs/apt` repo | Create (external) | reprepro config, KEY.gpg, .gitignore, GitHub Pages |
| `jmt-labs/homebrew-tap` repo | Create (external) | Empty public repo; GoReleaser pushes formula |

---

## Task 1: Prerequisites — External Repos, GPG Key, Secrets

**Files:** none (manual setup)

- [ ] **Step 1: Create `jmt-labs/homebrew-tap` on GitHub**

  Go to https://github.com/organizations/jmt-labs/repositories/new  
  Name: `homebrew-tap`, Visibility: Public, no template. Create empty repo.

- [ ] **Step 2: Create `jmt-labs/apt` on GitHub**

  Name: `apt`, Visibility: Public. Enable GitHub Pages: Settings → Pages → Source: `Deploy from a branch` → Branch: `gh-pages` / `/ (root)`.

- [ ] **Step 3: Generate GPG key pair for apt signing**

  ```bash
  gpg --batch --gen-key <<EOF
  Key-Type: RSA
  Key-Length: 4096
  Subkey-Type: RSA
  Subkey-Length: 4096
  Name-Real: jmt-labs
  Name-Email: markush1986@gmail.com
  Expire-Date: 0
  %no-protection
  %commit
  EOF

  # List the new key and note the KEY_ID (8-char hex after the slash)
  gpg --list-secret-keys --keyid-format=long
  # Example output:
  # sec   rsa4096/AABBCCDD11223344 2026-05-16 [SC]
  # KEY_ID = AABBCCDD11223344

  # Export public key for the apt repo
  gpg --armor --export AABBCCDD11223344 > /tmp/KEY.gpg

  # Export private key for CI
  gpg --armor --export-secret-keys AABBCCDD11223344 > /tmp/GPG_PRIVATE_KEY.txt
  ```

- [ ] **Step 4: Create GitHub PAT**

  GitHub → Settings → Developer settings → Personal access tokens → Fine-grained:
  - Repository access: `jmt-labs/homebrew-tap` + `jmt-labs/apt`
  - Permissions: Contents (read & write)
  
  Copy the token.

- [ ] **Step 5: Set CI Secrets in `jmt-labs/forgecrate`**

  GitHub → Settings → Secrets and variables → Actions → New repository secret:

  | Name | Value |
  |---|---|
  | `HOMEBREW_TAP_TOKEN` | PAT from Step 4 |
  | `GPG_PRIVATE_KEY` | Contents of `/tmp/GPG_PRIVATE_KEY.txt` |
  | `GPG_KEY_ID` | Key ID from Step 3 (e.g. `AABBCCDD11223344`) |
  | `CHOCOLATEY_API_KEY` | From chocolatey.org → Account → API Key |

- [ ] **Step 6: Create Chocolatey account**

  Register at https://community.chocolatey.org — confirm email, then copy API key from Account settings.

---

## Task 2: Initialize `jmt-labs/apt` Repository

**Files (in `jmt-labs/apt` repo):**
- Create: `conf/distributions`
- Create: `conf/options`
- Create: `.gitignore`
- Create: `KEY.gpg`
- Create: `README.md`

- [ ] **Step 1: Clone the new apt repo locally**

  ```bash
  git clone https://github.com/jmt-labs/apt.git /tmp/jmt-labs-apt
  cd /tmp/jmt-labs-apt
  git checkout --orphan gh-pages
  git rm -rf . 2>/dev/null || true
  ```

- [ ] **Step 2: Create reprepro config**

  ```bash
  mkdir -p conf
  cat > conf/distributions <<'EOF'
  Origin: jmt-labs
  Label: forgecrate
  Codename: stable
  Architectures: amd64 arm64
  Components: main
  Description: jmt-labs apt repository
  SignWith: AABBCCDD11223344
  EOF
  # Replace AABBCCDD11223344 with your actual GPG_KEY_ID from Task 1
  ```

  ```bash
  cat > conf/options <<'EOF'
  verbose
  basedir .
  EOF
  ```

- [ ] **Step 3: Add .gitignore**

  ```bash
  cat > .gitignore <<'EOF'
  db/
  EOF
  ```

- [ ] **Step 4: Add public GPG key**

  ```bash
  cp /tmp/KEY.gpg ./KEY.gpg
  ```

- [ ] **Step 5: Add minimal README**

  ```bash
  cat > README.md <<'EOF'
  # jmt-labs apt repository

  ```bash
  curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
    | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
  echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
    | sudo tee /etc/apt/sources.list.d/jmt-labs.list
  sudo apt update && sudo apt install forgecrate
  ```
  EOF
  ```

- [ ] **Step 6: Commit and push**

  ```bash
  git add .
  git commit -m "chore: initialize apt repository"
  git push origin gh-pages
  ```

---

## Task 3: GoReleaser Configuration

**Files (in `jmt-labs/forgecrate`):**
- Create: `.goreleaser.yaml`
- Create: `LICENSE`

- [ ] **Step 1: Create LICENSE file**

  ```bash
  cat > LICENSE <<'EOF'
  MIT License

  Copyright (c) 2026 Markus Hartmann

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
  EOF
  ```

- [ ] **Step 2: Create `.goreleaser.yaml`**

  ```yaml
  version: 2

  before:
    hooks:
      - go mod tidy

  builds:
    - main: ./cmd/forgecrate/
      binary: forgecrate
      env:
        - CGO_ENABLED=0
      goos:
        - linux
        - windows
        - darwin
      goarch:
        - amd64
        - arm64
      ignore:
        - goos: darwin
          goarch: amd64

  archives:
    - formats: [tar.gz]
      format_overrides:
        - goos: windows
          formats: [zip]
      name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

  checksum:
    name_template: "checksums.txt"

  nfpms:
    - package_name: forgecrate
      maintainer: Markus Hartmann <markush1986@gmail.com>
      description: Reproducible Claude Code setup for any repository.
      homepage: https://github.com/jmt-labs/forgecrate
      license: MIT
      formats:
        - deb
      bindir: /usr/local/bin

  brews:
    - name: forgecrate
      repository:
        owner: jmt-labs
        name: homebrew-tap
        token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
      homepage: https://github.com/jmt-labs/forgecrate
      description: Reproducible Claude Code setup for any repository.
      install: |
        bin.install "forgecrate"
      test: |
        system "#{bin}/forgecrate", "--version"

  chocolateys:
    - name: forgecrate
      title: forgecrate
      owners: jmt-labs
      description: Reproducible Claude Code setup for any repository.
      homepage: https://github.com/jmt-labs/forgecrate
      license_url: https://github.com/jmt-labs/forgecrate/blob/main/LICENSE
      require_license_acceptance: false
      api_key: "{{ .Env.CHOCOLATEY_API_KEY }}"
      source_repo: "https://push.chocolatey.org/"
      skip_publish: false
  ```

- [ ] **Step 3: Install GoReleaser locally and validate config**

  ```bash
  # macOS
  brew install goreleaser

  # Validate (does not build, just checks config)
  goreleaser check
  ```

  Expected output: `• release configuration is valid`

- [ ] **Step 4: Commit**

  ```bash
  git add .goreleaser.yaml LICENSE
  git commit -m "feat: add GoReleaser config with Homebrew, Chocolatey and deb support"
  ```

---

## Task 4: Update Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Replace the release target**

  Open `Makefile`. Replace:
  ```makefile
  release:
  	mkdir -p dist
  	GOOS=linux   GOARCH=amd64 go build -o dist/forgecrate-linux-amd64       ./cmd/forgecrate/
  	GOOS=linux   GOARCH=arm64 go build -o dist/forgecrate-linux-arm64       ./cmd/forgecrate/
  	GOOS=windows GOARCH=amd64 go build -o dist/forgecrate-windows-amd64.exe ./cmd/forgecrate/
  	GOOS=windows GOARCH=arm64 go build -o dist/forgecrate-windows-arm64.exe ./cmd/forgecrate/
  	GOOS=darwin  GOARCH=arm64 go build -o dist/forgecrate-darwin-arm64      ./cmd/forgecrate/
  ```

  With:
  ```makefile
  release:
  	goreleaser release --clean
  ```

- [ ] **Step 2: Verify build still works**

  ```bash
  make build
  ```

  Expected: `go build -o forgecrate ./cmd/forgecrate/` — no errors.

- [ ] **Step 3: Update clean target to remove goreleaser artifacts**

  The `clean` target already removes `dist/` — no change needed.

- [ ] **Step 4: Commit**

  ```bash
  git add Makefile
  git commit -m "feat: replace manual release target with goreleaser"
  ```

---

## Task 5: Update Release Workflow

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Rewrite `.github/workflows/release.yml`**

  Replace the entire file with:

  ```yaml
  name: Release

  on:
    workflow_dispatch:
      inputs:
        version:
          description: 'Version (z.B. v1.0.0)'
          required: true

  jobs:
    release:
      name: Release ${{ github.event.inputs.version }}
      runs-on: ubuntu-latest
      permissions:
        contents: write

      steps:
        - uses: actions/checkout@v4
          with:
            fetch-depth: 0

        - name: Release-Branch anlegen
          run: |
            git checkout -b releases/${{ github.event.inputs.version }}
            git push origin releases/${{ github.event.inputs.version }}

        - uses: actions/setup-go@v5
          with:
            go-version: '1.22'
            cache: true

        - run: make test
        - run: make test-e2e

        - name: Tag setzen
          run: |
            git tag ${{ github.event.inputs.version }}
            git push origin ${{ github.event.inputs.version }}

        - name: GoReleaser
          uses: goreleaser/goreleaser-action@v6
          with:
            version: latest
            args: release --clean
          env:
            GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
            HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
            CHOCOLATEY_API_KEY: ${{ secrets.CHOCOLATEY_API_KEY }}
            GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}
            GPG_KEY_ID: ${{ secrets.GPG_KEY_ID }}

    publish-apt:
      name: Publish apt repository
      needs: release
      runs-on: ubuntu-latest

      steps:
        - uses: actions/checkout@v4
          with:
            repository: jmt-labs/apt
            ref: gh-pages
            token: ${{ secrets.HOMEBREW_TAP_TOKEN }}

        - name: Install reprepro
          run: sudo apt-get install -y reprepro

        - name: Import GPG key
          run: |
            echo "${{ secrets.GPG_PRIVATE_KEY }}" | gpg --batch --import

        - name: Download .deb packages from release
          run: |
            mkdir -p /tmp/debs
            gh release download ${{ github.event.inputs.version }} \
              --repo jmt-labs/forgecrate \
              --pattern "*.deb" \
              --dir /tmp/debs
          env:
            GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

        - name: Add packages to apt repository
          run: |
            for deb in /tmp/debs/*.deb; do
              reprepro -b . includedeb stable "$deb"
            done

        - name: Commit and push
          run: |
            git config user.name "github-actions[bot]"
            git config user.email "github-actions[bot]@users.noreply.github.com"
            git add -A
            git commit -m "Release ${{ github.event.inputs.version }}" || echo "Nothing to commit"
            git push
  ```

- [ ] **Step 2: Validate workflow YAML syntax**

  ```bash
  # Requires actionlint (brew install actionlint)
  actionlint .github/workflows/release.yml
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  git add .github/workflows/release.yml
  git commit -m "feat: release workflow — GoReleaser + apt publishing job"
  ```

---

## Task 6: README Overhaul

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Replace the Installation section**

  Open `README.md`. Replace the entire `## Quick Start` section (lines 13–56) with:

  ```markdown
  ## Installation

  ### Homebrew (macOS / Linux)

  ```sh
  brew tap jmt-labs/tap
  brew install forgecrate
  ```

  ### Chocolatey (Windows)

  ```sh
  choco install forgecrate
  ```

  ### apt (Ubuntu / Debian)

  ```sh
  curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
    | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
  echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
    | sudo tee /etc/apt/sources.list.d/jmt-labs.list
  sudo apt update && sudo apt install forgecrate
  ```

  ### go install

  ```sh
  go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest
  ```

  ### curl (manual install, no package manager)

  ```sh
  curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash
  ```

  Specific version:

  ```sh
  curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash -s v1.0.0
  ```

  ---

  ## Quick Start

  Initialize a repository:

  ```sh
  forgecrate init --profile backend --flavors tdd
  ```

  This writes:

  ```
  CLAUDE.md · AGENTS.md · .claude/settings.json · .claude/commands/ · .claude/hooks/
  ```

  Update to the latest version:

  ```sh
  forgecrate update
  ```

  Switch profile:

  ```sh
  forgecrate update --profile fullstack
  ```
  ```

- [ ] **Step 2: Update the opening description to English**

  Replace the German paragraph after `# forgecrate`:

  From:
  ```markdown
  forgecrate deployt ein reproduzierbares Claude-Setup in beliebige Repos. Ein globales Go-Binary holt Konfiguration, Skills und Hooks von GitHub und compositioniert sie per Layer-System ins Ziel-Repo.
  ```

  To:
  ```markdown
  **forgecrate** deploys a reproducible [Claude Code](https://claude.ai/code) configuration to any repository. A single Go binary fetches profiles, skills, hooks, and MCP server definitions from GitHub and composes them into the target repo via a layered configuration system.
  ```

- [ ] **Step 3: Add version badge after the banner image**

  After the closing `</div>` tag (line 3), add:

  ```markdown
  [![Latest Release](https://img.shields.io/github/v/release/jmt-labs/forgecrate)](https://github.com/jmt-labs/forgecrate/releases/latest)
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add README.md
  git commit -m "docs: README overhaul — English, all install methods, version badge"
  ```

---

## Task 7: Create PR and verify

- [ ] **Step 1: Push branch and open PR**

  ```bash
  git push -u origin feat/package-distribution
  gh pr create \
    --title "feat: Homebrew, Chocolatey und apt distribution via GoReleaser" \
    --body "Closes #<issue-nr>

  ## Was
  - GoReleaser ersetzt manuelles make release
  - Homebrew formula via jmt-labs/homebrew-tap
  - Chocolatey community repo publishing
  - Ubuntu/Debian apt via GitHub Pages (jmt-labs/apt)
  - README komplett überarbeitet (Englisch, alle Installationswege)

  ## Wie getestet
  - goreleaser check: config valide
  - make build: Binary baut weiterhin
  - actionlint: Workflow YAML valide
  - Erster echter Release-Lauf verifiziert alle drei Kanäle"
  ```

- [ ] **Step 2: After merge — verify first real release**

  Trigger workflow manually with a test version (e.g. `v0.0.1-test`), then check:

  ```bash
  # Homebrew
  brew tap jmt-labs/tap && brew install forgecrate

  # Chocolatey (Windows VM oder CI)
  choco install forgecrate

  # apt
  curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
    | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
  echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
    | sudo tee /etc/apt/sources.list.d/jmt-labs.list
  sudo apt update && sudo apt install forgecrate
  forgecrate --version
  ```
