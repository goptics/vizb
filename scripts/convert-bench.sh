#!/usr/bin/env bash
set -euo pipefail

VIZB_ARGS=()
[ -n "${INPUT_TAG:-}" ] && VIZB_ARGS+=(--tag "$INPUT_TAG")
[ -n "${INPUT_NAME:-}" ] && VIZB_ARGS+=(-n "$INPUT_NAME")
[ -n "${INPUT_DESCRIPTION:-}" ] && VIZB_ARGS+=(-d "$INPUT_DESCRIPTION")
[ -n "${INPUT_GROUP_PATTERN:-}" ] && VIZB_ARGS+=(-p "$INPUT_GROUP_PATTERN")
[ -n "${INPUT_GROUP_REGEX:-}" ] && VIZB_ARGS+=(-r "$INPUT_GROUP_REGEX")
[ "${INPUT_SCALE:-}" != "linear" ] && VIZB_ARGS+=(--scale "$INPUT_SCALE")
[ -n "${INPUT_SORT:-}" ] && VIZB_ARGS+=(--sort "$INPUT_SORT")
[ -n "${INPUT_FILTER:-}" ] && VIZB_ARGS+=(--filter "$INPUT_FILTER")
[ -n "${INPUT_MEM_UNIT:-}" ] && VIZB_ARGS+=(-M "$INPUT_MEM_UNIT")
[ -n "${INPUT_TIME_UNIT:-}" ] && VIZB_ARGS+=(-T "$INPUT_TIME_UNIT")
[ -n "${INPUT_NUMBER_UNIT:-}" ] && VIZB_ARGS+=(-N "$INPUT_NUMBER_UNIT")
[ -n "${INPUT_CHARTS:-}" ] && VIZB_ARGS+=(-c "$INPUT_CHARTS")
[ "${INPUT_SHOW_LABELS:-}" = "true" ] && VIZB_ARGS+=(-l)

if [ "${INPUT_SOURCE:-}" = "file" ]; then
  cp "$INPUT_BENCH_FILE" bench-input.txt
else
  ${INPUT_BENCH_CMD} > bench-input.txt
fi

vizb bench-input.txt -o bench.json "${VIZB_ARGS[@]}"
