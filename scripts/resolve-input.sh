#!/usr/bin/env bash
set -euo pipefail

HAS_MERGE=""
[ -n "${INPUT_MERGE_FILES:-}" ] && HAS_MERGE="true"
[ -n "${INPUT_MERGE_DIR:-}" ] && HAS_MERGE="true"

if [ -n "${INPUT_BENCH_FILE:-}" ]; then
  echo "source=file" >> "$GITHUB_OUTPUT"
  echo "has_input=true" >> "$GITHUB_OUTPUT"
elif [ -n "${INPUT_BENCH_CMD:-}" ]; then
  echo "source=command" >> "$GITHUB_OUTPUT"
  echo "has_input=true" >> "$GITHUB_OUTPUT"
elif [ -n "$HAS_MERGE" ]; then
  echo "source=merge-only" >> "$GITHUB_OUTPUT"
  echo "has_input=false" >> "$GITHUB_OUTPUT"
else
  echo "::error::No input provided. Specify bench-cmd, bench-file, merge-files, or merge-dir."
  exit 1
fi
