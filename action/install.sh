#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="${HOME}/.local/bin"
mkdir -p "$BIN_DIR"

if [ "${RUNNER_OS:-}" = "Windows" ]; then
  DEST="${BIN_DIR}/vizb.exe"
else
  DEST="${BIN_DIR}/vizb"
fi

VIZB_BINARY="${VIZB_BINARY:-}"
if [ -n "$VIZB_BINARY" ]; then
  if [ ! -f "$VIZB_BINARY" ] && [ -f "${VIZB_BINARY}.exe" ]; then
    VIZB_BINARY="${VIZB_BINARY}.exe"
  fi
  echo "::notice::Installing local vizb binary from $VIZB_BINARY"
  cp "$VIZB_BINARY" "$DEST"
  chmod +x "$DEST"
  exit 0
fi

OS=$(echo "${RUNNER_OS:-}" | tr '[:upper:]' '[:lower:]')
[ "$OS" = "macos" ] && OS="darwin"
ARCH=$(echo "${RUNNER_ARCH:-}" | tr '[:upper:]' '[:lower:]')
[ "$ARCH" = "x64" ] && ARCH="amd64"

TAG="${VIZB_TAG:-}"
if [ -z "$TAG" ]; then
  echo "::error::VIZB_TAG is required when VIZB_BINARY is not set"
  exit 1
fi

VERSION="${TAG#v}"
EXT=".tar.gz"
[ "$OS" = "windows" ] && EXT=".zip"

URL="https://github.com/goptics/vizb/releases/download/${TAG}/vizb@${VERSION}-${OS}-${ARCH}${EXT}"

if [ "${VIZB_DRY_RUN:-}" = "1" ]; then
  echo "dest=${DEST}"
  echo "url=${URL}"
  exit 0
fi

curl -sfL "$URL" -o vizb-archive || {
  echo "::error::Failed to download vizb from $URL. Check that release ${TAG} exists."
  exit 1
}

if [ "$OS" = "windows" ]; then
  unzip -o vizb-archive -d "$BIN_DIR"
  if [ ! -f "${BIN_DIR}/vizb.exe" ] && [ -f "${BIN_DIR}/vizb" ]; then
    mv "${BIN_DIR}/vizb" "$DEST"
  fi
else
  tar xzf vizb-archive -C "$BIN_DIR"
  chmod +x "${BIN_DIR}/vizb"
fi
