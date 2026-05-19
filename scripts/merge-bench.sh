#!/usr/bin/env bash
set -euo pipefail

FILES=()
[ -f bench.json ] && FILES+=("bench.json")
[ -n "${INPUT_MERGE_FILES:-}" ] && FILES+=(${INPUT_MERGE_FILES})
[ -n "${INPUT_MERGE_DIR:-}" ] && FILES+=("${INPUT_MERGE_DIR}")

vizb merge "${FILES[@]}" -o bench.json --tag-axis "${INPUT_TAG_AXIS:-n}"
