#!/usr/bin/env bash
set -euo pipefail

REF="${GITHUB_ACTION_REF:-v0}"

if echo "$REF" | grep -qE '^v[0-9]+$'; then
  PREFIX="${REF#v}"
  TAG=$(git ls-remote --tags --refs --sort='-version:refname' \
    https://github.com/goptics/vizb "refs/tags/v${PREFIX}.*" |
    head -1 | sed 's|.*/||')
  if [ -z "$TAG" ]; then
    echo "::error::No release tags found for $REF"
    exit 1
  fi
else
  TAG="$REF"
fi

echo "tag=$TAG" >> "$GITHUB_OUTPUT"
