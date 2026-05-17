#!/usr/bin/env bash
# curl -fsSL https://raw.githubusercontent.com/VIGOR-Digital-Solution/vigor-cli/main/install.sh | bash
# Installs the latest `vigor` binary into /usr/local/bin (or $VIGOR_INSTALL_DIR).

set -euo pipefail

REPO="VIGOR-Digital-Solution/vigor-cli"
INSTALL_DIR="${VIGOR_INSTALL_DIR:-/usr/local/bin}"

# Detect OS + arch
case "$(uname -s)" in
  Linux*)   OS=linux ;;
  Darwin*)  OS=darwin ;;
  *) echo "Unsupported OS: $(uname -s) (use Homebrew or scoop on Windows)"; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "Unsupported arch: $(uname -m)"; exit 1 ;;
esac

# Get the latest version
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -n1 | sed -E 's/.*"([^"]+)".*/\1/')"
[[ -n "$TAG" ]] || { echo "Could not resolve latest release."; exit 1; }
VER="${TAG#v}"

URL="https://github.com/${REPO}/releases/download/${TAG}/vigor_${VER}_${OS}_${ARCH}.tar.gz"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading vigor ${TAG} for ${OS}/${ARCH}…"
curl -fsSL "$URL" -o "${TMP}/vigor.tar.gz"
tar -xzf "${TMP}/vigor.tar.gz" -C "$TMP"

if [[ -w "$INSTALL_DIR" ]]; then
  install -m 0755 "${TMP}/vigor" "${INSTALL_DIR}/vigor"
else
  echo "Installing into ${INSTALL_DIR} (needs sudo)…"
  sudo install -m 0755 "${TMP}/vigor" "${INSTALL_DIR}/vigor"
fi

echo ""
"${INSTALL_DIR}/vigor" --version
echo ""
echo "Installed. Try: vigor doctor"
