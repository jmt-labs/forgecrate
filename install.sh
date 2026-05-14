#!/usr/bin/env bash
set -euo pipefail

REPO="jmt-labs/claude-setup"
VERSION="${1:-latest}"

# OS erkennen
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  ;;
  darwin) ;;
  *)      echo "Nicht unterstützt: $OS" >&2; exit 1 ;;
esac

# Architektur erkennen
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             echo "Nicht unterstützt: $ARCH" >&2; exit 1 ;;
esac

# macOS: nur arm64 verfügbar — auf Intel via Rosetta 2 kompatibel
if [ "$OS" = "darwin" ] && [ "$ARCH" = "amd64" ]; then
  echo "Hinweis: Nur darwin/arm64 verfügbar — wird via Rosetta 2 ausgeführt." >&2
  ARCH="arm64"
fi

# Release-Metadaten holen
if [ "$VERSION" = "latest" ]; then
  API_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
  API_URL="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

PATTERN="claude-setup-${OS}-${ARCH}"
DOWNLOAD_URL=$(curl -fsSL "$API_URL" \
  | grep "browser_download_url" \
  | grep "$PATTERN" \
  | head -1 \
  | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Kein Binary gefunden für ${OS}/${ARCH} in Release '${VERSION}'" >&2
  exit 1
fi

# Installationspfad bestimmen
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Herunterladen und installieren
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

echo "Lade ${DOWNLOAD_URL##*/} herunter..."
curl -fsSL -o "$TMP" "$DOWNLOAD_URL"
chmod +x "$TMP"
mv "$TMP" "$INSTALL_DIR/claude-setup"
trap - EXIT

echo "Installiert: $INSTALL_DIR/claude-setup"

# Warnung wenn Verzeichnis nicht im PATH
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  echo ""
  echo "Hinweis: $INSTALL_DIR ist nicht in \$PATH."
  echo "Füge folgende Zeile zu deiner Shell-Konfiguration hinzu:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
