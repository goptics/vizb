#!/usr/bin/env bash
set -euo pipefail

REPO="goptics/vizb"
BIN="vizb"
INSTALL_DIR="/usr/local/bin"

# ----- helpers -----
die() { echo "error: $*" >&2; exit 1; }
info() { echo " info: $*"; }

# ----- detect platform -----
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  ;;
  darwin) ;;
  *) die "unsupported OS: $OS (only linux and macOS are supported)" ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) die "unsupported architecture: $ARCH" ;;
esac

# ----- check prerequisites -----
command -v curl >/dev/null 2>&1 || die "curl is required but not installed"
command -v tar >/dev/null 2>&1 || die "tar is required but not installed"

# ----- fetch latest version -----
info "fetching latest release…"
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [[ -z "$LATEST" ]]; then
  die "failed to determine latest version"
fi
VERSION="${LATEST#v}"
info "latest version: ${VERSION}"

# ----- download & extract -----
ARCHIVE="vizb@${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "downloading ${URL}…"
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}" || die "download failed"

info "extracting…"
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"
if [[ ! -f "${TMPDIR}/${BIN}" ]]; then
  die "binary not found in archive"
fi

# ----- install -----
info "installing to ${INSTALL_DIR}…"
if [[ -w "$INSTALL_DIR" ]]; then
  cp "${TMPDIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
else
  sudo cp "${TMPDIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
fi

chmod +x "${INSTALL_DIR}/${BIN}"

# ----- verify -----
info "verifying installation…"
if "${INSTALL_DIR}/${BIN}" --version >/dev/null 2>&1; then
  info "vizb ${VERSION} installed successfully to ${INSTALL_DIR}/${BIN}"
else
  die "installation verification failed"
fi
