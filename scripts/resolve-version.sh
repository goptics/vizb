#!/usr/bin/env bash
set -euo pipefail

REF="${INPUT_ACTION_REF:-}"
if [ -z "$REF" ] || [ "$REF" = "latest" ]; then
  PATH_REF=$(basename "$GITHUB_ACTION_PATH")
  if echo "$PATH_REF" | grep -qE '^v[0-9]'; then
    REF="$PATH_REF"
  fi
fi
if [ -z "$REF" ]; then
  REF="latest"
fi

if [ "$REF" = "latest" ]; then
  TAG=$(curl -sL -o /dev/null -w '%{url_effective}' \
    https://github.com/goptics/vizb/releases/latest | sed 's|.*/tag/||')
  if [ -z "$TAG" ] || [ "$TAG" = "latest" ]; then
    echo "::error::No releases found. Pin to a specific version tag."
    exit 1
  fi
  echo "tag=$TAG" >> "$GITHUB_OUTPUT"
  echo "is-latest=true" >> "$GITHUB_OUTPUT"
else
  echo "tag=$REF" >> "$GITHUB_OUTPUT"
  echo "is-latest=false" >> "$GITHUB_OUTPUT"
fi
