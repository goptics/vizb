#!/usr/bin/env bash
# Usage: curl https://vizb.goptics.org/install.sh | bash

set -euo pipefail

show_logo() {
cat <<'EOF'

    \    ####
     \ ++++        vizb installer
      \**   izb    vizb.goptics.org
       =

EOF
}
show_logo

REPO="goptics/vizb"
BIN="vizb"
INSTALL_DIR="${HOME}/.local/bin"

# ----- helpers -----
die() { printf "  \033[31m>\033[0m error: %s\n" "$*" >&2; exit 1; }
log() { printf "  \033[32m>\033[0m %s\n" "$*"; }

# ----- detect platform -----
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  ;;
  darwin) ;;
  *) die "unsupported OS: $OS" ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) die "unsupported architecture: $ARCH" ;;
esac

log "detected ${OS}/${ARCH}"

# ----- check prerequisites -----
command -v curl >/dev/null 2>&1 || die "curl is required but not installed"
command -v tar >/dev/null 2>&1 || die "tar is required but not installed"

# ----- fetch latest version -----
log "fetching latest release..."
LATEST=$(curl -fsSL -o /dev/null -w '%{url_effective}' \
  "https://github.com/${REPO}/releases/latest" \
  | grep -oP '[^/]+(?=/?$)')
if [[ -z "$LATEST" ]]; then
  die "failed to determine latest version"
fi
VERSION="${LATEST#v}"

# ----- download & extract -----
ARCHIVE="vizb@${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

log "downloading v${VERSION}..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}" || die "download failed"

log "extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"
if [[ ! -f "${TMPDIR}/${BIN}" ]]; then
  die "binary not found in archive"
fi

# ----- install -----
mkdir -p "$INSTALL_DIR"

cp "${TMPDIR}/${BIN}" "${INSTALL_DIR}/${BIN}"
chmod +x "${INSTALL_DIR}/${BIN}"

# ----- verify -----
if ! "${INSTALL_DIR}/${BIN}" --version >/dev/null 2>&1; then
  die "installation verification failed"
fi

log "installed vizb to ${INSTALL_DIR}/${BIN}"
echo

if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  log "note: add ${INSTALL_DIR} to your PATH"
  log "      echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
fi

log "ready. run 'vizb' to get started"
