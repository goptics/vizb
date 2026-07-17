#!/usr/bin/env bash
# Maintainer helpers for CHANGELOG.md drafts and release preflight.
# Usage:
#   ./scripts/release-changelog.sh draft   vX.Y.Z
#   ./scripts/release-changelog.sh commit  vX.Y.Z
#   ./scripts/release-changelog.sh preflight vX.Y.Z
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

usage() {
  echo "Usage: $0 {draft|commit|preflight} vX.Y.Z" >&2
  exit 1
}

require_version() {
  local version="${1:-}"
  if [[ -z "$version" ]]; then
    echo "Usage: $0 $cmd vX.Y.Z" >&2
    exit 1
  fi
  if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "VERSION must look like vX.Y.Z (got: $version)" >&2
    exit 1
  fi
}

has_section() {
  grep -qE "^# \\[${1}\\]" CHANGELOG.md
}

cmd="${1:-}"
version="${2:-}"
[[ -n "$cmd" ]] || usage

case "$cmd" in
  draft)
    require_version "$version"
    command -v git-cliff >/dev/null 2>&1 || {
      echo "git-cliff not found. Install: brew install git-cliff  (or: task cliff:install)" >&2
      exit 1
    }
    [[ -f cliff.toml ]] || {
      echo "cliff.toml not found in repo root" >&2
      exit 1
    }
    if has_section "$version"; then
      echo "CHANGELOG.md already has a section for ${version}; refusing to prepend again." >&2
      exit 1
    fi
    git-cliff --config cliff.toml --unreleased --tag "$version" --prepend CHANGELOG.md
    cat <<EOF

Draft prepended for ${version}.
Next:
  1. Polish CHANGELOG.md if needed
  2. task release:changelog:commit -- ${version}
  3. task release -- ${version}
EOF
    ;;

  commit)
    require_version "$version"
    if ! has_section "$version"; then
      echo "CHANGELOG.md has no section for ${version}. Run: task release:changelog -- ${version}" >&2
      exit 1
    fi
    if git diff --quiet -- CHANGELOG.md && git diff --cached --quiet -- CHANGELOG.md; then
      echo "No CHANGELOG.md changes to commit." >&2
      exit 1
    fi
    git add CHANGELOG.md
    git commit -m "docs(changelog): add ${version} release notes"
    echo "Committed CHANGELOG.md for ${version}. Push, then: task release -- ${version}"
    ;;

  preflight)
    require_version "$version"
    if ! has_section "$version"; then
      echo "CHANGELOG.md has no section for ${version}." >&2
      echo "Run: task release:changelog -- ${version}, polish, commit, then retry." >&2
      exit 1
    fi
    if ! git diff --quiet -- CHANGELOG.md || ! git diff --cached --quiet -- CHANGELOG.md; then
      echo "CHANGELOG.md has uncommitted changes. Commit first:" >&2
      echo "  task release:changelog:commit -- ${version}" >&2
      exit 1
    fi
    ;;

  *)
    usage
    ;;
esac
