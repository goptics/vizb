#!/usr/bin/env bash
set -euo pipefail

OS=$(echo "$INPUT_RUNNER_OS" | tr '[:upper:]' '[:lower:]')
[ "$OS" = "macos" ] && OS="darwin"
ARCH=$(echo "$INPUT_RUNNER_ARCH" | tr '[:upper:]' '[:lower:]')
[ "$ARCH" = "x64" ] && ARCH="amd64"

TAG="$INPUT_TAG"
VERSION="${TAG#v}"
EXT=".tar.gz"
[ "$OS" = "windows" ] && EXT=".zip"

URL="https://github.com/goptics/vizb/releases/download/${TAG}/vizb@${VERSION}-${OS}-${ARCH}${EXT}"

mkdir -p ~/.local/bin
curl -sfL "$URL" -o vizb-archive
if [ $? -ne 0 ]; then
  echo "::error::Failed to download vizb from $URL. Check that release ${TAG} exists."
  exit 1
fi

if [ "$OS" = "windows" ]; then
  unzip vizb-archive -d ~/.local/bin
else
  tar xzf vizb-archive -C ~/.local/bin
  chmod +x ~/.local/bin/vizb
fi

echo "$HOME/.local/bin" >> "$GITHUB_PATH"
